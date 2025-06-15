package queries

const CreateBalance = `
	INSERT INTO balances (user_id)
	VALUES ($1)
`

const GetBalanceByUserID = `
	SELECT * FROM balances WHERE user_id = $1;
`
