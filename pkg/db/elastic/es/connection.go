package es

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"net/http"
	"time"

	es "github.com/elastic/go-elasticsearch/v7"
)

type Info struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	ClusterUUID string `json:"cluster_uuid"`
	Version     struct {
		Number                           string    `json:"number"`
		BuildFlavor                      string    `json:"build_flavor"`
		BuildType                        string    `json:"build_type"`
		BuildHash                        string    `json:"build_hash"`
		BuildDate                        time.Time `json:"build_date"`
		BuildSnapshot                    bool      `json:"build_snapshot"`
		LuceneVersion                    string    `json:"lucene_version"`
		MinimumWireCompatibilityVersion  string    `json:"minimum_wire_compatibility_version"`
		MinimumIndexCompatibilityVersion string    `json:"minimum_index_compatibility_version"`
	} `json:"version"`
	Tagline string `json:"tagline"`
}

type Connection struct {
	ctx context.Context
	db  *es.Client
	cfg *env.Elastic
}

func NewConnection(ctx context.Context, p *env.Elastic) *Connection {
	return &Connection{ctx: ctx, cfg: p}
}

func (r *Connection) DB() *es.Client {
	return r.db
}

// Connect initializes db connetion
func (r *Connection) Connect() error {
	retryBackoff := backoff.NewExponentialBackOff()
	retryBackoff.InitialInterval = time.Second

	cl, err := es.NewClient(es.Config{
		Addresses:            r.cfg.Hosts,
		Username:             r.cfg.User,
		Password:             r.cfg.Pass,
		EnableRetryOnTimeout: r.cfg.EnableRetryOnTimeout,
		MaxRetries:           r.cfg.MaxRetries,
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			d := retryBackoff.NextBackOff()
			log.InfoF(r.ctx, "Attempt: %d | Sleeping for %s...\n", i, d)
			return d
		},
		CompressRequestBody:     r.cfg.CompressRequestBody,
		EnableCompatibilityMode: r.cfg.EnableCompatibilityMode,
		EnableDebugLogger:       r.cfg.Debug,
		EnableMetrics:           r.cfg.EnableMetrics,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   r.cfg.MaxIdleConnsPerHost,
			ResponseHeaderTimeout: time.Duration(r.cfg.ResponseHeaderTimeoutSec) * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
		Logger: &CJSONLogger{ctx: r.ctx},
	})

	if err != nil {
		return cerror.New(r.ctx, cerror.ElasticToKind(err), err).LogError()
	}

	res, err := cl.Info()
	if err != nil {
		return cerror.New(r.ctx, cerror.KindElasticOther, err).LogError()
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			_ = cerror.NewF(r.ctx, cerror.KindElasticOther, "close response es info error: %+v", err).LogError()
		}
	}()

	var info Info
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return cerror.New(r.ctx, cerror.KindElasticOther, err).LogError()
	}

	r.logInfo(&info)

	r.db = cl

	return nil
}

// Close closes db connetion
func (r *Connection) Close() error {
	if success, err := r.ensureConnection(); !success {
		return err
	}

	return nil
}

func (r *Connection) logInfo(info *Info) {
	event := logger.NewEvent(r.ctx, logger.LevelInfo, "elasticsearch client")
	values := make(map[string]interface{})

	if info != nil {
		values["version"] = info.Version.Number
		values["lucene_version"] = info.Version.LuceneVersion
		values["name"] = info.Name
		values["cluster"] = info.ClusterName
		values["cluster_uuid"] = info.ClusterUUID
		values["build_flavor"] = info.Version.BuildFlavor
		values["build_type"] = info.Version.BuildType
		values["build_hash"] = info.Version.BuildHash
		values["build_date"] = info.Version.BuildDate
		values["build_snapshot"] = info.Version.BuildSnapshot
		values["min_wire_version"] = info.Version.MinimumWireCompatibilityVersion
		values["min_index_version"] = info.Version.MinimumIndexCompatibilityVersion
		values["tagline"] = info.Tagline
	}

	for key, val := range values {
		event.WithValue(key, val)
	}

	log.Log(event)
}

func (r *Connection) ensureConnection() (bool, error) {
	isConnected := r.db != nil
	if !isConnected {
		return false, cerror.NewF(r.ctx, cerror.KindElasticOther, "db connection is not initialized").LogError()
	}

	return true, nil
}
