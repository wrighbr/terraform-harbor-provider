package provider

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/bestsellerit/terraform-provider-harbor/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var pathRegistry = "/api/registries"

type credential struct {
	AccessKey    string `json:"access_key,omitempty"`
	AccessSecret string `json:"access_secret,omitempty"`
	Type         string `json:"type,omitempty"`
}

type registry struct {
	Credential  credential `json:"credential"`
	ID          int        `json:"id,omitempty"`
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	Insecure    bool       `json:"insecure"`
	Type        string     `json:"type"`
	Description string     `json:"description,omitempty"`

	// The below is used to update registry
	AccessKey      string `json:"access_key,omitempty"`
	AccessSecret   string `json:"access_secret,omitempty"`
	CredentialType string `json:"credential_type,omitempty"`
}

func resourceRegistry() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				// ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  false,
			},
			"provider_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"url_endpoint": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"access_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access_secret": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
		Create: resourceRegistryCreate,
		Read:   resourceRegistryRead,
		Update: resourceRegistryUpdate,
		Delete: resourceRegistryDelete,
	}
}

func resourceRegistryCreate(d *schema.ResourceData, m interface{}) error {
	apiClient, body := registryBody(d, m)

	_, err := apiClient.SendRequest("POST", pathRegistry, body, 201)
	if err != nil {
		return err
	}

	return resourceRegistryRead(d, m)
}

func resourceRegistryRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	resp, err := apiClient.SendRequest("GET", pathRegistry+"?name="+d.Get("name").(string), nil, 200)

	var jsonData []registry
	err = json.Unmarshal([]byte(resp), &jsonData)
	if err != nil {
		return fmt.Errorf("[ERROR] Unable to unmarchal: %s", err)
	}

	d.Set("description", jsonData[0].Description)
	d.Set("name", jsonData[0].Name)
	d.SetId(strconv.Itoa(jsonData[0].ID))

	return nil
}

func resourceRegistryUpdate(d *schema.ResourceData, m interface{}) error {
	apiClient, body := registryBody(d, m)
	body.AccessKey = d.Get("access_id").(string)
	body.AccessSecret = d.Get("access_secret").(string)

	_, err := apiClient.SendRequest("PUT", pathRegistry+"/"+d.Id(), body, 200)
	if err != nil {
		return err
	}
	return resourceRegistryRead(d, m)
}

func resourceRegistryDelete(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	url := pathRegistry + "/" + d.Id()
	apiClient.SendRequest("DELETE", url, nil, 200)
	return nil
}

func registryBody(d *schema.ResourceData, m interface{}) (*client.Client, registry) {
	apiClient := m.(*client.Client)

	creds := credential{
		AccessKey:    d.Get("access_id").(string),
		AccessSecret: d.Get("access_secret").(string),
		Type:         "basic",
	}
	body := registry{
		Name:        d.Get("name").(string),
		URL:         d.Get("url_endpoint").(string),
		Insecure:    false,
		Type:        d.Get("provider_type").(string),
		Description: d.Get("description").(string),
		Credential:  creds,
	}

	return apiClient, body
}
