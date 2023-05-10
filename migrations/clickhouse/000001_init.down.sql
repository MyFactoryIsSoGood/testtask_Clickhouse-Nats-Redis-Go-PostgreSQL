ALTER TABLE default.Items DROP INDEX idx_items_name;
ALTER TABLE default.Items DROP INDEX idx_items_campaign_id;
ALTER TABLE default.Items DROP INDEX idx_items_id;

DROP TABLE Items;
