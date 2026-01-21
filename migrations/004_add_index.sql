-- delivery: быстрый GetByOrderID/DeleteByOrderID
CREATE UNIQUE INDEX IF NOT EXISTS ux_delivery_order_id
ON delivery(order_id);

-- delivery: для ReleaseExpired / join по courier_id
CREATE INDEX IF NOT EXISTS ix_delivery_deadline
ON delivery(deadline);

CREATE INDEX IF NOT EXISTS ix_delivery_courier_id
ON delivery(courier_id);
