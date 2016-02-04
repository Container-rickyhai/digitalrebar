package main

/* These are needed for database access
 *   "database/sql"
 *    _ "github.com/lib/pq"
 *
 * These are needed for consul api accces
 *    "github.com/hashicorp/consul/api"
 */

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/digitalrebar/gcfg"
)

// For NSUPDATE, Dns.Server is ip of server to update
// For PDNS, Dns.Server to access (localhost)
// For BIND, Dns.Server name (FQDN of DNS server)
type Config struct {
	Dns struct {
		Type     string
		Hostname string
		Password string
		Port     int
		Server   string
	}
	Network struct {
		Port     int
		Username string
		Password string
	}
}

var config_path, key_pem, cert_pem, data_dir, backingStore string

func init() {
	flag.StringVar(&config_path, "config_path", "/etc/dns-mgmt.conf", "Path to config file")
	flag.StringVar(&key_pem, "key_pem", "/etc/dns-mgmt-https-key.pem", "Path to config file")
	flag.StringVar(&cert_pem, "cert_pem", "/etc/dns-mgmt-https-cert.pem", "Path to config file")
	flag.StringVar(&data_dir, "data_dir", "/var/cache/rebar-dns-mgmt", "Path to store data")
	flag.StringVar(&backingStore, "backing_store", "file", "Backing store to use. Either 'consul' or 'file'")
}

func main() {
	flag.Parse()

	var cfg Config
	err := gcfg.ReadFileInto(&cfg, config_path)
	if err != nil {
		log.Fatal(err)
	}

	var be dns_backend_point

	if cfg.Dns.Type == "BIND" {
		be = NewBindDnsInstance(cfg.Dns.Server)
	} else if cfg.Dns.Type == "POWERDNS" {
		base := fmt.Sprintf("http://%s:%d/servers/%s", cfg.Dns.Hostname, cfg.Dns.Port, cfg.Dns.Server)
		be = &PowerDnsInstance{
			UrlBase:     base,
			AccessToken: cfg.Dns.Password,
		}
	} else if cfg.Dns.Type == "NSUPDATE" {
		be = NewNsupdateDnsInstance(cfg.Dns.Server)
	} else {
		log.Fatal("Failed to find type")
	}
	var bs LoadSaver
	switch backingStore {
	case "file":
		bs, err = NewFileStore(data_dir + "/database.json")
	case "consul":
		bs, err = NewConsulStore(data_dir)
	default:
		log.Fatalf("Unknown backing store type %s", backingStore)
	}

	fe := NewFrontend(&be, bs)

	fe.load_data()

	api := rest.NewApi()
	api.Use(&rest.AuthBasicMiddleware{
		Realm: "test zone",
		Authenticator: func(userId string, password string) bool {
			if userId == cfg.Network.Username && password == cfg.Network.Password {
				return true
			}
			return false
		},
	})
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/zones", fe.GetAllZones),
		rest.Get("/zones/#id", fe.GetZone),
		&rest.Route{"PATCH", "/zones/#id", fe.PatchZone},
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	connStr := fmt.Sprintf(":%d", cfg.Network.Port)
	log.Println("Using", connStr)

	log.Fatal(http.ListenAndServeTLS(connStr, cert_pem, key_pem, api.MakeHandler()))
}
