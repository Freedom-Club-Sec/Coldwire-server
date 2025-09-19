package config

import (
    "encoding/json"
    "strings"
    "fmt"
    "os"
    "errors"

    "github.com/Freedom-Club-Sec/Coldwire-server/internal/utils"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/crypto"
    "github.com/Freedom-Club-Sec/Coldwire-server/internal/constants"
)

type redisConfig struct {
    Host     string
    Password string
    Port     uint16
    DB       uint16 

}

type sqlConfig struct {
    Host        string
    Port        uint16 
    DBName      string `json:"db_name"`
    DBUser      string `json:"db_user"`
    DBPassword  string `json:"db_password"`
}

type Config struct {
    DomainOrIP         string      `json:"Your_domain_or_IP"`
    FederationEnabled  bool        `json:"Federation_enabled"`
    UserStorage        string      `json:"User_storage"`
    DataStorage        string      `json:"Data_storage"`
    Redis              redisConfig `json:"Redis"`
    SQL                sqlConfig   `json:"SQL"`
    BlacklistedDomains []string    `json:"Blacklisted_Domain_Names"`
    BlacklistedIPs     []string    `json:"Blacklisted_IP_nets"`
    JWTSecret          []byte      `json:"JWT_Secret_Base64_Encoded"`
    DSAPrivateKey      []byte      `json:"ML_DSA_87_Private_Key_Base64_Encoded"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    err = json.Unmarshal(data, &cfg)
    if err != nil {
        return nil, err
    }

    // Enforce lowercase for less error-prone code.
    cfg.DomainOrIP = strings.ToLower(cfg.DomainOrIP)
    cfg.UserStorage = strings.ToLower(cfg.UserStorage)
    cfg.DataStorage = strings.ToLower(cfg.DataStorage)

    // Sanity check the configuration
    err = cfg.Validate()
    if err != nil {
        return nil, err
    }

    if cfg.JWTSecret == nil || len(cfg.JWTSecret) == 0 {
        cfg.JWTSecret, err = utils.SecureRandomBytes(constants.JWT_SECRET_LEN)
        if err != nil {
            return nil, err
        }

        cfg.Write(path)
    }

    if cfg.DSAPrivateKey == nil || len(cfg.DSAPrivateKey) == 0 {
        _, privateKey, err := crypto.CreateDSAKeyPair()
        if err != nil {
            return nil, err
        }

        cfg.DSAPrivateKey, err = privateKey.MarshalBinary()
        if err != nil {
            return nil, err
        }
        cfg.Write(path)
    }

    return &cfg, nil
}

func (c *Config) Write(path string) error {
    jsonBytes, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(path, jsonBytes, 0644)
}

func (c *Config) Validate() error {
    switch c.UserStorage {
    case "sql", "internal":
    default:
        return fmt.Errorf("Invalid user storage mechanism: %s", c.UserStorage)
    }

    switch c.DataStorage {
    case "internal", "redis", "sql":
    default:
        return fmt.Errorf("Invalid data storage:  %s", c.UserStorage)
    }

    if c.Redis.Port == 0 {
        return fmt.Errorf("Invalid Redis port: %d", c.Redis.Port)
    }

    if c.SQL.Port == 0 {
        return fmt.Errorf("Invalid SQL port: %d", c.SQL.Port)
    }

    if len(c.DomainOrIP) == 0 {
        return errors.New("You must include your domain name or IP address in the configuration file.")
    }

    return nil
}
