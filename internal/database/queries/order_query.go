package queries

const CreateOrder = `
	INSERT INTO orders (numeral_id, user_id, status, accrual) 
	VALUES ($1, $2, $3, $4);
`

const GetOrderByNumeralID = `
	SELECT * FROM orders WHERE numeral_id = $1;
`

const GetOrdersByUserID = `
	SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC;
`

const UpdateOrder = `
	UPDATE orders 
	SET status = $1, accrual = $2, updated_at = NOW()
	WHERE numeral_id = $3
`
