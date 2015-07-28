package comm

var err error

err = RunIfNoError(err, func () {
    return nil
})

err = RunIfNoError(err, func () {
    return errors.New("abc")
})

err = RunIfNoError(err, func () {
    return errors.New("ddd")
})


func RunIfNoError(err error, callback func()) error {
	if err != nil {
		return err
	}
	return callback()
}
