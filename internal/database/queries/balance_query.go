package queries

const CreateBalance = `
	INSERT INTO balances (user_id)
	VALUES ($1)
`

const GetBalanceByUserID = `
	SELECT * FROM balances WHERE user_id = $1;
`

const UpdateBalance = `
	UPDATE balances 
	SET current = $1, updated_at = CURRENT_TIMESTAMP 
	WHERE id = $2;
`
