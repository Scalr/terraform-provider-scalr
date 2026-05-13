package client

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/disco"

	"github.com/scalr/go-scalr"

	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
	clientV2 "github.com/scalr/go-scalr/v2/scalr/client"

	"github.com/scalr/terraform-provider-scalr/internal/logging"
)

const (
	discoveryMaxRetries = 5
	discoveryMaxDelay   = 120 * time.Second
)

var scalrServiceIDs = []string{"iacp.v3"}

// Configure configures and returns a new Scalr client.
func Configure(h, t, v string) (*scalr.Client, *scalrV2.Client, error) {
	// Parse the hostname for comparison
	hostname, err := svchost.ForComparison(h)
	if err != nil {
		return nil, nil, err
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

	// Discover the address, retrying on transient network errors.
	host, err := discoverWithRetry(services, hostname)
	if err != nil {
		return nil, nil, err
	}

	// Get the full service address.
	var address *url.URL
	var discoErr error
	for _, scalrServiceID := range scalrServiceIDs {
		service, err := host.ServiceURL(scalrServiceID)
		if _, ok := err.(*disco.ErrVersionNotSupported); !ok && err != nil {
			return nil, nil, err
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
		return nil, nil, discoErr
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
		return nil, nil, errors.New("required token could not be found")
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
		return nil, nil, err
	}

	scalrClient.RetryServerErrors(true)

	// Client v2
	scalrClientV2 := scalrV2.NewClient(
		h,
		t,
		clientV2.WithRetryServerErrors(true),
		clientV2.WithAppInfo("terraform-provider-scalr", v),
	)

	return scalrClient, scalrClientV2, nil
}

// discoverWithRetry calls services.Discover and retries on transient network
// errors (timeouts, connection resets) with an exponential backoff.
// This mostly addresses network issues in the k8s environment during acceptance tests.
func discoverWithRetry(services *disco.Disco, hostname svchost.Hostname) (*disco.Host, error) {
	var (
		host  *disco.Host
		err   error
		delay = 5 * time.Second
	)
	for attempt := 0; attempt <= discoveryMaxRetries; attempt++ {
		host, err = services.Discover(hostname)
		if err == nil || !isTransientDiscoveryError(err) {
			break
		}
		if attempt < discoveryMaxRetries {
			log.Printf(
				"[WARN] Discovery attempt %d/%d failed: %s. Retrying in %s...",
				attempt+1, discoveryMaxRetries, err, delay,
			)
			time.Sleep(delay)
			delay *= 2
			if delay > discoveryMaxDelay {
				delay = discoveryMaxDelay
			}
		}
	}
	return host, err
}

// isTransientDiscoveryError returns true for network errors that are worth retrying.
func isTransientDiscoveryError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "Client.Timeout exceeded") ||
		strings.Contains(msg, "request canceled") ||
		strings.Contains(msg, "context deadline exceeded")
}
