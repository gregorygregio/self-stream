package dtaccess

type DbNotFound struct{}

func (e *DbNotFound) Error() string {
	return "not found"
}

type DbConnectionError struct{}

func (e *DbConnectionError) Error() string {
	return "database connection error"
}

type DbError struct{}

func (e *DbError) Error() string {
	return "database error"
}
