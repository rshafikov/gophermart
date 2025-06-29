package queries

const GetWdsByUserID = `
	SELECT * FROM withdrawals WHERE user_id = $1;
`

const GetWdsByBalanceID = `
	SELECT * FROM withdrawals WHERE balance_id = $1;
`

const SortByCreatedAt = ` ORDER BY created_at DESC`

const CreateWd = `
	INSERT INTO withdrawals (user_id, balance_id, order_numeral_id, amount, created_at)
	VALUES ($1, $2, $3, $4, $5);
`
