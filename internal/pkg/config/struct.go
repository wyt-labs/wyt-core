package config

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

type Duration time.Duration

func (d *Duration) MarshalText() (text []byte, err error) {
	return []byte(time.Duration(*d).String()), nil
}

func (d *Duration) UnmarshalText(b []byte) error {
	x, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

func (d *Duration) ToDuration() time.Duration {
	return time.Duration(*d)
}

func (d *Duration) String() string {
	return time.Duration(*d).String()
}

func StringToTimeDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(Duration(5)) {
			return data, nil
		}

		d, err := time.ParseDuration(data.(string))
		if err != nil {
			return nil, err
		}
		return Duration(d), nil
	}
}

type Email struct {
	UnsubscribeFrontendURL string `mapstructure:"unsubscribe_frontend_url" toml:"unsubscribe_frontend_url"`
	SenderAddress          string `mapstructure:"sender_address" toml:"sender_address"`
	SenderName             string `mapstructure:"sender_name" toml:"sender_name"`
	SenderPwd              string `mapstructure:"sender_pwd" toml:"sender_pwd"`
	MailServerHost         string `mapstructure:"mail_server_host" toml:"mail_server_host"`
	MailServerPort         int    `mapstructure:"mail_server_port" toml:"mail_server_port"`
	EnableTLS              bool   `mapstructure:"enable_tls" toml:"enable_tls"`
}

type App struct {
	NodeIndex     uint16        `mapstructure:"node_index" toml:"node_index"`
	AccessDomain  string        `mapstructure:"access_domain" toml:"access_domain"`
	AdminAddr     string        `mapstructure:"admin_addr" toml:"admin_addr"`
	HotTagNum     uint64        `mapstructure:"hot_tag_num" toml:"hot_tag_num"`
	HotTrackNum   uint64        `mapstructure:"hot_track_num" toml:"hot_track_num"`
	RetryTime     int           `mapstructure:"retry_time" toml:"retry_time"`
	RetryInterval Duration      `mapstructure:"retry_interval" toml:"retry_interval"`
	Email         Email         `mapstructure:"email" toml:"email"`
	CaculateLimit CaculateLimit `mapstructure:"caculate_limit" toml:"caculate_limit"`
}

type CaculateLimit struct {
	FinancingAmountLimit uint64 `mapstructure:"financing_amount_limit" toml:"financing_amount_limit"`
	FinancingTimeLimit   int64  `mapstructure:"financing_time_limit" toml:"financing_time_limit"`
}

type HTTP struct {
	Port                  int      `mapstructure:"port" toml:"port"`
	MultipartMemory       int64    `mapstructure:"-" toml:"multipart_memory"`
	ReadTimeout           Duration `mapstructure:"read_timeout" toml:"read_timeout"`
	WriteTimeout          Duration `mapstructure:"write_timeout" toml:"write_timeout"`
	TLSEnable             bool     `mapstructure:"tls_enable" toml:"tls_enable"`
	TLSCertFilePath       string   `mapstructure:"tls_cert_file_path" toml:"tls_cert_file_path"`
	TLSKeyFilePath        string   `mapstructure:"tls_key_file_path" toml:"tls_key_file_path"`
	JWTTokenValidDuration Duration `mapstructure:"jwt_token_valid_duration" toml:"jwt_token_valid_duration"`
	JWTTokenHMACKey       string   `mapstructure:"jwt_token_hmac_key" toml:"jwt_token_hmac_key"`
}

type DBInfo struct {
	IP       string `mapstructure:"ip" toml:"ip"`
	Port     uint32 `mapstructure:"port" toml:"port"`
	Username string `mapstructure:"username" toml:"username"`
	Password string `mapstructure:"password" toml:"password"`
	DBName   string `mapstructure:"db_name" toml:"db_name"`
}

type Mongodb struct {
	DBInfo          `mapstructure:",squash"`
	IsSrv           bool     `mapstructure:"is_srv" toml:"is_srv"`
	ConnectTimeout  Duration `mapstructure:"connect_timeout" toml:"connect_timeout"`
	MaxPoolSize     int      `mapstructure:"max_pool_size" toml:"max_pool_size"`
	MaxConnIdleTime Duration `mapstructure:"max_conn_idle_time" toml:"max_conn_idle_time"`
}

type DB struct {
	Service Mongodb `mapstructure:"service" toml:"service"`
}

type Backends struct {
	MetabaseURL      string `mapstructure:"metabase_url" toml:"metabase_url"`
	MetabaseUserName string `mapstructure:"metabase_username" toml:"metabase_username"`
	MetabasePassword string `mapstructure:"metabase_password" toml:"metabase_password"`
}

type DatasourceCoincap struct {
	APIEndpoint        string `mapstructure:"api_endpoint" toml:"api_endpoint"`
	WebsocketEndpoint  string `mapstructure:"websocket_endpoint" toml:"websocket_endpoint"`
	APIKey             string `mapstructure:"api_key" toml:"api_key"`
	SingleWsTokenLimit int    `mapstructure:"single_ws_token_limit" toml:"single_ws_token_limit"`
}

type DatasourceBinance struct {
	SingleWsTokenLimit int `mapstructure:"single_ws_token_limit" toml:"single_ws_token_limit"`
}

type DatasourceOkx struct {
	APIEndpoint        string `mapstructure:"api_endpoint" toml:"api_endpoint"`
	WebsocketEndpoint  string `mapstructure:"websocket_endpoint" toml:"websocket_endpoint"`
	SingleWsTokenLimit int    `mapstructure:"single_ws_token_limit" toml:"single_ws_token_limit"`
}

type DatasourceTokenTerminal struct {
	APIEndpoint string `mapstructure:"api_endpoint" toml:"api_endpoint"`
	APIKey      string `mapstructure:"api_key" toml:"api_key"`
}

type DatasourceCmc struct {
	APIEndpoint string `mapstructure:"api_endpoint" toml:"api_endpoint"`
	ApiKey      string `mapstructure:"api_key" toml:"api_key"`
}

type Datasource struct {
	Market Market `mapstructure:"market" toml:"market"`
	Metric Metric `mapstructure:"metric" toml:"metric"`
	Social Social `mapstructure:"social" toml:"social"`
}

type Market struct {
	Disable                        bool              `mapstructure:"disable" toml:"disable"`
	MarketDataRefreshInterval      Duration          `mapstructure:"market_data_refresh_interval" toml:"market_data_refresh_interval"`
	Last7DaysKlinesDataRefreshCron string            `mapstructure:"last_7_days_klines_data_refresh_cron" toml:"last_7_days_klines_data_refresh_cron"`
	MarketDrivers                  []string          `mapstructure:"market_drivers" toml:"market_drivers"`
	Coincap                        DatasourceCoincap `mapstructure:"coincap" toml:"coincap"`
	Binance                        DatasourceBinance `mapstructure:"binance" toml:"binance"`
	Okx                            DatasourceOkx     `mapstructure:"okx" toml:"okx"`
	Cmc                            DatasourceCmc     `mapstructure:"cmc" toml:"cmc"`
}
type Metric struct {
	Disable                   bool                    `mapstructure:"disable" toml:"disable"`
	ActiveUserDataRefreshCron string                  `mapstructure:"active_user_data_refresh_cron" toml:"active_user_data_refresh_cron"`
	TokenTerminal             DatasourceTokenTerminal `mapstructure:"token_terminal" toml:"token_terminal"`
}

type Social struct {
	Disable bool    `mapstructure:"disable" toml:"disable"`
	Github  Github  `mapstructure:"github" toml:"github"`
	Twitter Twitter `mapstructure:"twitter" toml:"twitter"`
}

type Github struct {
	RetryLimit int `mapstructure:"retry_limit" toml:"retry_limit"`
	// 并发数
	ConcurrencyLimit int      `mapstructure:"concurrency_limit" toml:"concurrency_limit"`
	RefreshCron      string   `mapstructure:"refresh_cron" toml:"refresh_cron"`
	HttpTimeout      Duration `mapstructure:"http_timeout" toml:"http_timeout"`
	APIKeys          []string `mapstructure:"api_keys" toml:"api_keys"`
	RefreshInterval  Duration `mapstructure:"refresh_interval" toml:"refresh_interval"`
	Checkpoint       int      `mapstructure:"checkpoint" toml:"checkpoint"`
	CacheTimeout     Duration `mapstructure:"cache_timeout" toml:"cache_timeout"`
}

type Twitter struct {
	RetryLimit  int      `mapstructure:"retry_limit" toml:"retry_limit"`
	Checkpoint  int      `mapstructure:"checkpoint" toml:"checkpoint"`
	RefreshCron string   `mapstructure:"refresh_cron" toml:"refresh_cron"`
	HttpTimeout Duration `mapstructure:"http_timeout" toml:"http_timeout"`
}

type Chatgpt struct {
	Endpoint        string  `mapstructure:"endpoint" toml:"endpoint"`
	EndpointFull    string  `mapstructure:"endpoint_full" toml:"endpoint_full"`
	APIKey          string  `mapstructure:"api_key" toml:"api_key"`
	Model           string  `mapstructure:"model" toml:"model"`
	Temperature     float32 `mapstructure:"temperature" toml:"temperature"`
	PresencePenalty float32 `mapstructure:"presence_penalty" toml:"presence_penalty"`
}

type Extension struct {
	Chatgpt Chatgpt `mapstructure:"chatgpt" toml:"chatgpt"`
}

type Cache struct {
	ExpiredTime     Duration `mapstructure:"expired_time" toml:"expired_time"`
	CleanupInterval Duration `mapstructure:"cleanup_interval" toml:"cleanup_interval"`
}

type Log struct {
	Level        string   `mapstructure:"level" toml:"level"`
	Filename     string   `mapstructure:"file_name" toml:"file_name"`
	MaxAge       Duration `mapstructure:"max_age" toml:"max_age"`
	MaxSizeStr   string   `mapstructure:"max_size" toml:"max_size"`
	MaxSize      int64    `mapstructure:"-" toml:"-"`
	RotationTime Duration `mapstructure:"rotation_time" toml:"rotation_time"`
}

type ETL struct {
	DataRefreshCron string `mapstructure:"data_refresh_cron" toml:"data_refresh_cron"`
	Coincarp        DB     `mapstructure:"coincarp" toml:"coincarp"`
}

type AIBackend struct {
	AIEnv           string `mapstructure:"ai_env" toml:"ai_env"`
	Endpoint        string `mapstructure:"end_point" toml:"end_point"`
	APIKey          string `mapsctructure:"api_key" toml:"api_key"`
	ProjectId       string `mapstructure:"project_id" toml:"project_id"`
	UniProjectId    string `mapstructure:"uni_project_id" toml:"uni_project_id"`
	DevEndpoint     string `mapstructure:"dev_endpoint" toml:"dev_endpoint"`
	DevAPIKey       string `mapstructure:"dev_api_key" toml:"dev_api_key"`
	DevProjectId    string `mapstructure:"dev_project_id" toml:"dev_project_id"`
	DevUniProjectId string `mapstructure:"dev_uni_project_id" toml:"dev_uni_project_id"`
}

type Okx struct {
	Endpoint   string `mapstructure:"endpoint" toml:"endpoint"`
	ProjectId  string `mapstructure:"project_id" toml:"project_id"`
	APIKey     string `mapstructure:"api_key" toml:"api_key"`
	SecretKey  string `mapstructure:"secret_key" toml:"secret_key"`
	Passphrase string `mapstructure:"passphrase" toml:"passphrase"`
}

type Config struct {
	RootPath string `mapstructure:"-" toml:"-"`
	App      App    `mapstructure:"app" toml:"app"`

	HTTP       HTTP       `mapstructure:"http" toml:"http"`
	DB         DB         `mapstructure:"db" toml:"db"`
	Backends   Backends   `mapstructure:"backends" toml:"backends"`
	AIBackend  AIBackend  `mapstructure:"ai_backend" toml:"ai_backend"`
	Okx        Okx        `mapstructure:"okx" toml:"okx"`
	Datasource Datasource `mapstructure:"datasource" toml:"datasource"`
	Extension  Extension  `mapstructure:"extension" toml:"extension"`
	Cache      Cache      `mapstructure:"cache" toml:"cache"`
	Log        Log        `mapstructure:"log" toml:"log"`
	ETL        ETL        `mapstructure:"etl" toml:"etl"`
}
