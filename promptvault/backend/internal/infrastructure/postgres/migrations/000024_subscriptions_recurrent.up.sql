-- Recurrent payments support для T-Bank автопродления.
-- rebill_id выдаётся T-Bank при первом успешном Init с Recurrent=Y;
-- используется при Charge для последующих списаний без 3DS.
-- auto_renew управляет включён ли автоматический ререкур: при false renewal loop
-- пропускает подписку (юзер сам платит вручную).
ALTER TABLE subscriptions ADD COLUMN rebill_id  VARCHAR(50);
ALTER TABLE subscriptions ADD COLUMN auto_renew BOOLEAN NOT NULL DEFAULT TRUE;

-- Индекс для эффективной выборки подписок к продлению (за 3 дня до окончания).
CREATE INDEX idx_subscriptions_renewal
    ON subscriptions (status, auto_renew, current_period_end)
    WHERE status = 'active' AND auto_renew = TRUE AND rebill_id IS NOT NULL;
