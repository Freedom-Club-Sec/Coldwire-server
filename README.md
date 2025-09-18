# Coldwire-server
Coldwire federated server implementation in Golang. 

# Configuration
Copy default configuration file from `configs/config.json`, then add your domain name, or IP to `Your_domain_or_IP`

By default configuration uses "internal" for database store, which uses sqlite3 database. No configuration needed.

If you want better scalibility, and performence, we highly recommend setting `User_storage` to `SQL`, then configuring the `SQL` block with your database creditntials
And set `Data_storage` to either `SQL`, or "Redis".



# Usage
```bash
go build cmd/server/main.go
```

Then running:
```bash
./main -c your_config_path
```
