package utils

func ReturnNilError(err error) (error, error) {

	if err != nil {
		return nil, err
	}
	return nil, nil
}
