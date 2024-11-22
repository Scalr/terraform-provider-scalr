package client

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/logging"
)

var scalrServiceIDs = []string{"iacp.v3"}

func Configure(h, t, v string) (*scalr.Client, error) {
	// Parse the hostname for comparison,
	hostname, err := svchost.ForComparison(h)
	if err != nil {
		return nil, err
	}

	providerUaString := fmt.Sprintf("terraform-provider-scalr/%s", v)

	// Get the Terraform CLI configuration.
	config := cliConfig()

	// Create a new credential source and service discovery object.
	credsSrc := credentialsSource(config)
	services := disco.NewWithCredentialsSource(credsSrc)
	services.SetUserAgent(providerUaString)
	services.Transport = logging.NewLoggingTransport(services.Transport)

	// Add any static host configurations service discovery object.
	for userHost, hostConfig := range config.Hosts {
		host, err := svchost.ForComparison(userHost)
		if err != nil {
			// ignore invalid hostnames.
			continue
		}
		services.ForceHostServices(host, hostConfig.Services)
	}

	// Discover the address.
	host, err := services.Discover(hostname)
	if err != nil {
		return nil, err
	}

	// Get the full service address.
	var address *url.URL
	var discoErr error
	for _, scalrServiceID := range scalrServiceIDs {
		service, err := host.ServiceURL(scalrServiceID)
		if _, ok := err.(*disco.ErrVersionNotSupported); !ok && err != nil {
			return nil, err
		}
		// If discoErr is nil we save the first error. When multiple services
		// are checked, and we found one that didn't give an error we need to
		// reset the discoErr. So if err is nil, we assign it as well.
		if discoErr == nil || err == nil {
			discoErr = err
		}
		if service != nil {
			address = service
			break
		}
	}

	// When we don't have any constraints errors, also check for discovery
	// errors before we continue.
	if discoErr != nil {
		return nil, discoErr
	}

	// Only try to get to the token from the credentials source if no token
	// was explicitly set in the provider configuration.
	if t == "" {
		creds, err := services.CredentialsForHost(hostname)
		if err != nil {
			log.Printf("[DEBUG] Failed to get credentials for %s: %s (ignoring)", hostname, err)
		}
		if creds != nil {
			t = creds.Token()
		}
	}

	// If we still don't have a token at this point, we return an error.
	if t == "" {
		return nil, errors.New("required token could not be found")
	}

	httpClient := scalr.DefaultConfig().HTTPClient
	httpClient.Transport = logging.NewLoggingTransport(httpClient.Transport)

	headers := make(http.Header)
	headers.Add("User-Agent", providerUaString)

	// Create a new Scalr client config
	cfg := &scalr.Config{
		Address:    address.String(),
		Token:      t,
		HTTPClient: httpClient,
		Headers:    headers,
	}

	// Create a new Scalr client.
	scalrClient, err := scalr.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	scalrClient.RetryServerErrors(true)
	return scalrClient, nil
}
