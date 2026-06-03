# Stripe Metadata Reference

All metadata fields set on Stripe objects by Confire.
Keep this in sync when adding new products or checkout flows.

---

## Products

### Confire Dev

| Key | Value |
|---|---|
| `product_kind` | `subscription` |
| `plan_id` | `dev` |
| `plan_name` | `Dev` |

### Confire Credit Pack

| Key | Value |
|---|---|
| `product_kind` | `topup` |
| `topup_id` | `remote_optimization_pack` |
| `credit_type` | `remote_optimization` |
| `credits_per_unit` | `5000` |

---

## Prices

### Dev monthly ($10/month)

| Key | Value |
|---|---|
| `product_kind` | `subscription` |
| `plan_id` | `dev` |
| `plan_name` | `Dev` |
| `billing_interval` | `monthly` |
| `credits_per_period` | `5000` |
| `credit_type` | `remote_optimization` |

### Dev annual ($90/year)

| Key | Value |
|---|---|
| `product_kind` | `subscription` |
| `plan_id` | `dev` |
| `plan_name` | `Dev` |
| `billing_interval` | `annual` |
| `credits_per_period` | `60000` |
| `credits_per_month` | `5000` |
| `credit_type` | `remote_optimization` |

### Credit Pack price ($5 one-time)

| Key | Value |
|---|---|
| `product_kind` | `topup` |
| `topup_id` | `remote_optimization_pack` |
| `credit_type` | `remote_optimization` |
| `credits_per_unit` | `5000` |
| `max_quantity` | `20` |

---

## Checkout Session metadata

### Subscription checkout

| Key | Value |
|---|---|
| `kind` | `subscription` |
| `plan_id` | `dev` |
| `billing_interval` | `monthly` or `annual` |
| `credit_type` | `remote_optimization` |

### Top-up checkout

| Key | Value |
|---|---|
| `kind` | `topup` |
| `topup_id` | `remote_optimization_pack` |
| `credit_type` | `remote_optimization` |
| `credits_per_unit` | `5000` |
| `max_quantity` | `20` |
| `quantity` | *(initial quantity, 1–20 — webhook re-reads from line_items)* |

---

## subscription_data metadata (subscription checkouts only)

Same as the checkout session metadata above — duplicated onto the subscription so webhook events on the subscription object carry the same fields.

| Key | Value |
|---|---|
| `kind` | `subscription` |
| `plan_id` | `dev` |
| `billing_interval` | `monthly` or `annual` |
| `credit_type` | `remote_optimization` |

---

## Stripe object IDs (dev environment)

| Object | ID |
|---|---|
| Dev product | `prod_UdYngLggCJUx5B` |
| Dev monthly price | `price_1TeHfTCjsgbilsT2kKli4wU5` |
| Dev annual price | `price_1TeHggCjsgbilsT2a56012vt` |
| Credit Pack product | `prod_UdYvfOnfvYjr9H` |
| Credit Pack price | `price_1TeHo1CjsgbilsT2qq4wmI2W` |

Production IDs go here once created.
