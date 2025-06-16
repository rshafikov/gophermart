package queries

const GetTxsByUserID = `
	SELECT * FROM transactions WHERE user_id = $1;
`

const GetTxsByBalanceID = `
	SELECT * FROM transactions WHERE balance_id = $1;
`

const SortByCreatedAt = ` ORDER BY created_at DESC`

const CreateTx = `
	INSERT INTO transactions (user_id, balance_id, order_numeral_id, amount, created_at)
	VALUES ($1, $2, $3, $4, $5);
`
