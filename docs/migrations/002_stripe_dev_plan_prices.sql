-- Migration 002 — Wire Stripe product/price IDs for Dev plans
-- Run in: Supabase Dashboard → SQL Editor
-- Apply to: dev project AND production project separately
--
-- Dev plan (monthly)
--   product: prod_UdYngLggCJUx5B
--   price:   price_1TeHfTCjsgbilsT2kKli4wU5
--
-- Dev plan (annual)
--   product: prod_UdYow0riWjmnFH
--   price:   price_1TeHggCjsgbilsT2a56012vt

UPDATE plans
SET config = jsonb_set(jsonb_set(config,
  '{stripe,productId}', '"prod_UdYngLggCJUx5B"'),
  '{stripe,priceId}',   '"price_1TeHfTCjsgbilsT2kKli4wU5"')
WHERE id = 'dev';

UPDATE plans
SET config = jsonb_set(jsonb_set(config,
  '{stripe,productId}', '"prod_UdYow0riWjmnFH"'),
  '{stripe,priceId}',   '"price_1TeHggCjsgbilsT2a56012vt"')
WHERE id = 'dev_annual';

-- Verify
SELECT id,
       config->'stripe'->>'productId' AS stripe_product_id,
       config->'stripe'->>'priceId'   AS stripe_price_id
FROM plans
WHERE id IN ('dev', 'dev_annual');
