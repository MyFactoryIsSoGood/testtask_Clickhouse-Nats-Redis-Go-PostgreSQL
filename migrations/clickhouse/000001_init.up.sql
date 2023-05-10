CREATE TABLE Items (
                       Id Int32,
                       CampaignId Int32,
                       Name String,
                       Description Nullable(String),
                       Priority Nullable(Int32),
                       Removed Nullable(Bool),
                       EventTime Nullable(DateTime)
) ENGINE = MergeTree()
ORDER BY Id;

ALTER TABLE default.Items ADD INDEX idx_items_id(Id) TYPE minmax GRANULARITY 8192;
ALTER TABLE default.Items ADD INDEX idx_items_campaign_id(CampaignId) TYPE minmax GRANULARITY 8192;
ALTER TABLE default.Items ADD INDEX idx_items_name(Name) TYPE bloom_filter GRANULARITY 8192;