-- TraceCore ClickHouse Schema
-- Database: tracecore

-- Traces Table: Stores individual spans from distributed traces
CREATE TABLE IF NOT EXISTS traces (
    trace_id String,
    span_id String,
    parent_span_id String,
    service_name String,
    operation_name String,
    start_time DateTime64(9),
    end_time DateTime64(9),
    duration_ns UInt64,
    status_code Enum8('unset'=0, 'ok'=1, 'error'=2),
    resource_id String,
    attributes String,  -- JSON string of span attributes
    date Date DEFAULT toDate(start_time)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, service_name, start_time, trace_id, span_id)
TTL date + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;

-- Service Map Table: Stores aggregated service-to-service dependencies with metrics
CREATE TABLE IF NOT EXISTS service_map (
    from_service String,
    to_service String,
    window_start DateTime,
    window_end DateTime,
    latency_p50 UInt32,   -- Latency in milliseconds
    latency_p95 UInt32,
    latency_p99 UInt32,
    request_count UInt64,
    error_count UInt64,
    error_rate Float32,
    request_rate Float32,  -- Requests per second
    sample_traces Array(String),  -- Sample trace IDs for debugging
    date Date DEFAULT toDate(window_start)
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, from_service, to_service, window_start)
TTL date + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Service Nodes Table: Stores service metadata and resource mappings
CREATE TABLE IF NOT EXISTS service_nodes (
    service_name String,
    service_type String,
    resource_id String,
    resource_type String,
    endpoints Array(String),
    first_seen DateTime,
    last_seen DateTime,
    metadata String,  -- JSON string of additional metadata
    date Date DEFAULT toDate(last_seen)
) ENGINE = ReplacingMergeTree(last_seen)
PARTITION BY toYYYYMM(date)
ORDER BY (service_name, resource_id, date)
TTL date + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Resource Service Mapping Table: Links infrastructure resources to services
CREATE TABLE IF NOT EXISTS resource_service_mappings (
    resource_id String,
    resource_type String,
    service_name String,
    confidence Float32,  -- Confidence score (0.0 to 1.0)
    first_seen DateTime,
    last_seen DateTime,
    trace_count UInt64,
    date Date DEFAULT toDate(last_seen)
) ENGINE = ReplacingMergeTree(last_seen)
PARTITION BY toYYYYMM(date)
ORDER BY (resource_id, service_name, date)
TTL date + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Indexes for better query performance

-- Index for trace ID lookups
ALTER TABLE traces ADD INDEX IF NOT EXISTS idx_trace_id trace_id TYPE bloom_filter GRANULARITY 1;

-- Index for service name lookups
ALTER TABLE traces ADD INDEX IF NOT EXISTS idx_service service_name TYPE bloom_filter GRANULARITY 1;

-- Index for operation name lookups
ALTER TABLE traces ADD INDEX IF NOT EXISTS idx_operation operation_name TYPE bloom_filter GRANULARITY 1;

-- Index for resource ID lookups
ALTER TABLE traces ADD INDEX IF NOT EXISTS idx_resource_id resource_id TYPE bloom_filter GRANULARITY 1;

-- Materialized views for common queries

-- Daily trace statistics by service
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_daily_trace_stats
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, service_name)
POPULATE AS
SELECT
    toDate(start_time) AS date,
    service_name,
    count() AS span_count,
    countIf(status_code = 'error') AS error_count,
    quantile(0.50)(duration_ns) AS p50_duration,
    quantile(0.95)(duration_ns) AS p95_duration,
    quantile(0.99)(duration_ns) AS p99_duration
FROM traces
GROUP BY date, service_name;

-- Hourly service map metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_hourly_service_edges
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, hour, from_service, to_service)
POPULATE AS
SELECT
    toDate(start_time) AS date,
    toHour(start_time) AS hour,
    parent.service_name AS from_service,
    child.service_name AS to_service,
    count() AS request_count,
    countIf(child.status_code = 'error') AS error_count,
    quantile(0.50)(child.duration_ns) AS p50_latency,
    quantile(0.95)(child.duration_ns) AS p95_latency,
    quantile(0.99)(child.duration_ns) AS p99_latency
FROM traces AS child
INNER JOIN traces AS parent ON child.parent_span_id = parent.span_id AND child.trace_id = parent.trace_id
GROUP BY date, hour, from_service, to_service;
