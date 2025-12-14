-- DeepDrift ClickHouse Schema
-- Database for storing drift events and impact analysis results

CREATE DATABASE IF NOT EXISTS deepdrift;

USE deepdrift;

-- Drift Events Table
-- Stores all drift events detected by TFDrift-Falco
CREATE TABLE IF NOT EXISTS drift_events (
    -- Primary identifiers
    id String,
    resource_id String,
    resource_type String,

    -- Drift metadata
    drift_type Enum8('created' = 1, 'modified' = 2, 'deleted' = 3),
    severity Enum8('low' = 1, 'medium' = 2, 'high' = 3, 'critical' = 4),

    -- Timestamps
    timestamp DateTime64(3),
    detected_at DateTime64(3) DEFAULT now64(),

    -- State snapshots (JSON)
    state_before String,  -- JSON string
    state_after String,   -- JSON string
    diff String,          -- JSON string

    -- Root cause information
    cloudtrail_event_id String,
    event_name String,
    user_identity String,
    user_arn String,
    source_ip String,
    root_cause_timestamp DateTime64(3),

    -- Metadata
    tags Map(String, String),

    -- Partitioning and ordering
    date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, resource_type, severity, timestamp)
TTL date + INTERVAL 90 DAY  -- Keep data for 90 days
SETTINGS index_granularity = 8192;

-- Impact Analysis Results Table
-- Stores impact analysis results for drift events
CREATE TABLE IF NOT EXISTS impact_analysis (
    -- Link to drift event
    drift_event_id String,

    -- Impact metrics
    affected_resource_count UInt32,
    blast_radius UInt8,
    severity Enum8('low' = 1, 'medium' = 2, 'high' = 3, 'critical' = 4),

    -- Analysis timestamp
    analyzed_at DateTime64(3) DEFAULT now64(),

    -- Affected resources (JSON array)
    affected_resources String,  -- JSON string

    -- Recommendations (JSON array)
    recommendations String,  -- JSON string

    -- Metadata
    date Date DEFAULT toDate(analyzed_at)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, severity, analyzed_at)
TTL date + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Affected Resources Detail Table
-- Normalized table for efficient querying of affected resources
CREATE TABLE IF NOT EXISTS affected_resources (
    drift_event_id String,
    resource_id String,
    resource_type String,
    relation_type String,  -- network, dependency, ownership
    distance UInt8,
    impact_description String,

    date Date DEFAULT today()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, resource_type, drift_event_id)
TTL date + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Materialized Views for Analytics

-- Daily drift summary
CREATE MATERIALIZED VIEW IF NOT EXISTS drift_summary_daily
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, resource_type, drift_type, severity)
AS SELECT
    toDate(timestamp) as date,
    resource_type,
    drift_type,
    severity,
    count() as count,
    countIf(cloudtrail_event_id != '') as with_root_cause
FROM drift_events
GROUP BY date, resource_type, drift_type, severity;

-- Hourly drift summary for real-time monitoring
CREATE MATERIALIZED VIEW IF NOT EXISTS drift_summary_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (datetime_hour, resource_type, severity)
AS SELECT
    toStartOfHour(timestamp) as datetime_hour,
    toDate(timestamp) as date,
    resource_type,
    drift_type,
    severity,
    count() as count
FROM drift_events
GROUP BY datetime_hour, date, resource_type, drift_type, severity;

-- User activity summary
CREATE MATERIALIZED VIEW IF NOT EXISTS user_activity_summary
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, user_identity, severity)
AS SELECT
    toDate(timestamp) as date,
    user_identity,
    severity,
    count() as drift_count,
    uniqExact(resource_id) as unique_resources
FROM drift_events
WHERE user_identity != ''
GROUP BY date, user_identity, severity;

-- Resource type impact summary
CREATE MATERIALIZED VIEW IF NOT EXISTS resource_impact_summary
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, resource_type)
AS SELECT
    toDate(de.timestamp) as date,
    de.resource_type,
    count() as drift_count,
    avg(ia.affected_resource_count) as avg_affected_resources,
    avg(ia.blast_radius) as avg_blast_radius,
    countIf(ia.severity = 'critical') as critical_count
FROM drift_events de
LEFT JOIN impact_analysis ia ON de.id = ia.drift_event_id
GROUP BY date, de.resource_type;

-- Indexes for faster queries

-- Full-text search index on resource_id
-- ALTER TABLE drift_events ADD INDEX resource_id_idx resource_id TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 4;

-- Index on user_identity for quick user lookup
-- ALTER TABLE drift_events ADD INDEX user_identity_idx user_identity TYPE bloom_filter() GRANULARITY 4;

-- Index on severity for filtering
-- ALTER TABLE drift_events ADD INDEX severity_idx severity TYPE set(10) GRANULARITY 4;

-- Common Queries Examples

-- Recent critical drifts
-- SELECT
--     timestamp,
--     resource_id,
--     drift_type,
--     user_identity
-- FROM drift_events
-- WHERE severity = 'critical'
--   AND timestamp > now() - INTERVAL 24 HOUR
-- ORDER BY timestamp DESC
-- LIMIT 100;

-- Drift trends over last 7 days
-- SELECT
--     date,
--     drift_type,
--     count() as count
-- FROM drift_summary_daily
-- WHERE date >= today() - INTERVAL 7 DAY
-- GROUP BY date, drift_type
-- ORDER BY date, drift_type;

-- Top users causing drifts
-- SELECT
--     user_identity,
--     sum(drift_count) as total_drifts,
--     sum(unique_resources) as unique_resources
-- FROM user_activity_summary
-- WHERE date >= today() - INTERVAL 30 DAY
-- GROUP BY user_identity
-- ORDER BY total_drifts DESC
-- LIMIT 20;

-- Resources with highest blast radius
-- SELECT
--     de.resource_id,
--     de.resource_type,
--     ia.affected_resource_count,
--     ia.blast_radius,
--     de.timestamp
-- FROM drift_events de
-- JOIN impact_analysis ia ON de.id = ia.drift_event_id
-- WHERE de.timestamp > now() - INTERVAL 7 DAY
-- ORDER BY ia.blast_radius DESC, ia.affected_resource_count DESC
-- LIMIT 50;
