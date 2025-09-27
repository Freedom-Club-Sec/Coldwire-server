# Coldwire-server
[![Tests](https://github.com/Freedom-Club-Sec/Coldwire-server/actions/workflows/tests.yml/badge.svg)](https://github.com/Freedom-Club-Sec/Coldwire-server/actions/workflows/tests.yml) [![Coverage](https://goreportcard.com/badge/github.com/Freedom-Club-Sec/Coldwire-server)](https://goreportcard.com/report/github.com/Freedom-Club-Sec/Coldwire-server)

Coldwire federated server implementation in Golang. 

# Configuration
Copy the example configuration from: [docs/example_configuration.json](https://github.com/Freedom-Club-Sec/Coldwire-server/blob/main/docs/example_config.json)

Put either your server DNS name or IP in `Your_domain_or_IP`

That's it! No further modification needed

If you want better performance, you might want configure "`SQL`" and or "`Redis`" services, 
consult [docs/configuration.md for more details](https://github.com/Freedom-Club-Sec/Coldwire-server/blob/main/docs/configuration.md)


# Setup

Download prebuilt binaries from releases.

Or optionally compile the source code using:
```bash
git clone https://github.com/Freedom-Club-Sec/Coldwire-server.git
cd Coldwire-server
make build
```
The compiled binary will be in `bin/` folder.

# Example Usage

```bash
./coldwire-server-linux-amd64 --help
Usage of ./coldwire-server-linux-amd64:
  -c string
        Path to JSON configuration file (default "configs/config.json")
  -h string
        Server address to listen on (default "127.0.0.1")
  -p int
        Server port to listen on (default 8000)
```


Run server:

```bash
./coldwire-server-linux-amd64 -c Your_Config_File.json
```
