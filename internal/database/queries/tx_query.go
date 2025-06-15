package queries

const GetTxsByUserID = `
	SELECT * FROM transactions WHERE user_id = $1;
`

const GetTxsByBalanceID = `
	SELECT * FROM transactions WHERE balance_id = $1;
`

const SortByCreatedAt = ` ORDER BY created_at DESC`
