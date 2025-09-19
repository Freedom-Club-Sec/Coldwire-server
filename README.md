# Coldwire-server
Coldwire federated server implementation in Golang. 

# Configuration
Copy the example configuration from from [docs/example_configuration.json](https://github.com/Freedom-Club-Sec/Coldwire-server/blob/main/docs/example_config.json)

Put either your server DNS name or IP in `Your_domain_or_IP`
That's it. No further modification needed

If you want better performance, you might want configure "`SQL`" and or "`Redis`" services: consult [docs/configuration.md for more details](https://github.com/Freedom-Club-Sec/Coldwire-server/blob/main/docs/configuration.md)


# Usage
```bash
go build cmd/server/main.go
```

Then running:
```bash
./main -c your_config_path
```
