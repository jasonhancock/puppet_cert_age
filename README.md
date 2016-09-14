# puppet_cert_age

Parses all signed certificates from the Puppetmaster's CA and sends the age of each certificate (in seconds) to a Graphite server via Carbon.

## Usage

```
$ ./puppet_cert_age -h
Usage of ./puppet_cert_age:
  -cert_dir string
        Puppet certificate directory (default "/etc/puppetlabs/puppet/ssl/ca/signed")
  -graphite_host string
        host and port of carbon server (default "127.0.0.1:2003")
  -prefix string
        metric prefix (default "servers")
  -suffix string
        metric prefix (default "puppet.cert_age")

More information at https://github.com/jasonhancock/puppet_cert_age
```

If you have a certificate with a common name of `client.example.com` and use the default options, the metric recorded will be `servers.client_example_com.puppet.cert_age`.
