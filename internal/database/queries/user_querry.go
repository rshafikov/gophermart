package queries

const CreateUser = `
	INSERT INTO users (login, password) 
	VALUES ($1, $2);
`

const GetUserByLogin = `
	SELECT * FROM users WHERE login = $1;
`
