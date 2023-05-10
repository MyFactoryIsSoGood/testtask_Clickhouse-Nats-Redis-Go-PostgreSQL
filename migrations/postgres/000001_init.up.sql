CREATE TABLE campaigns (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255)
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    campaign_id INT,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255),
    priority INT,
    removed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (campaign_id) REFERENCES campaigns (id)
);

CREATE INDEX idx_campaigns_id ON campaigns (id);
CREATE INDEX idx_items_id ON items (id);
CREATE INDEX idx_items_campaign_id ON items (campaign_id);

INSERT INTO campaigns (name) VALUES ('FAANG');

