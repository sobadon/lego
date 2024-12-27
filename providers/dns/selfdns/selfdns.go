// original: https://github.com/masa23/lego/blob/3b9bdfb360ae0307e8997a738a28a924d7e1ec1f/providers/dns/selfdns/selfdns.go

// Package self hosted DNS provider
package selfdns

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/miekg/dns"
)

var (
	envNamespace = "SELFDNS_"

	EnvListenAddress  = envNamespace + "LISTEN_ADDRESS"
	EnvServerHostname = envNamespace + "SERVER_HOSTNAME"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ListenAddress  string
	ServerHostname string
	fqdn           string
	value          string
	serverUDP      dns.Server
	serverTCP      dns.Server
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		ListenAddress:  env.GetOrDefaultString(EnvListenAddress, "0.0.0.0"),
		ServerHostname: env.GetOrDefaultString(EnvServerHostname, ""),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for self hosted DNS.
// The DNS server hostname and the Listen IP address are specified in the environment variables:
// SELFDNS_LISTEN_ADDRESS & SELFDNS_SERVER_HOSTNAME.
func NewDNSProvider() (*DNSProvider, error) {
	var err error
	values, err := env.Get(EnvListenAddress, EnvServerHostname)
	if err != nil {
		return nil, fmt.Errorf("selfdns: %w", err)
	}

	config := NewDefaultConfig()
	config.ListenAddress = values[EnvListenAddress]
	config.ServerHostname = values[EnvServerHostname]

	// If hostname is empty, get hostname and use it.
	if config.ServerHostname == "" {
		config.ServerHostname, err = os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("selfdns: %w", err)
		}
	}

	// Change format for IPv6 addresses
	if net.ParseIP(config.ListenAddress).To4() == nil {
		config.ListenAddress = "[" + config.ListenAddress + "]" // IPv6
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for self hosted DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("selfdns: the configuration of the DNS provider is nil")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	d.config.fqdn, d.config.value = dns01.GetRecord(domain, keyAuth)
	return d.Run(d.config.fqdn, d.config.value)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return d.Stop(domain, keyAuth)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval int) {
	return 120, 2
}
