package apperrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New("message")
	assert.Equal(t, "message", err.Error())

	appErr := Unwrap(err)
	assert.Equal(t, err.Error(), appErr.Err.Error())
	assert.Equal(t, "", appErr.Message)
}

func TestErrorf(t *testing.T) {
	err := Errorf("message %d", 123)
	assert.Equal(t, "message 123", err.Error())

	appErr := Unwrap(err)
	assert.Equal(t, err.Error(), appErr.Err.Error())
	assert.Equal(t, "", appErr.Message)
}

func TestWithMessage(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := WithMessage(nil, "message")
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("original")

		err1 := WithMessage(err0, "message")
		assert.Equal(t, "message", err1.Error())

		appErr := Unwrap(err1)
		assert.Equal(t, err0, appErr.Err)
		assert.Equal(t, err1.Error(), appErr.Message)
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("original")

		err1 := &Error{
			Err:        err0,
			Message:    "message 1",
			StatusCode: 400,
		}
		err2 := WithMessage(err1, "message 2")
		assert.Equal(t, "message 2", err2.Error())

		{
			appErr := Unwrap(err1)
			assert.Equal(t, err0, appErr.Err)
			assert.Equal(t, err1.Error(), appErr.Message)
			assert.Equal(t, 400, appErr.StatusCode)
		}

		{
			appErr := Unwrap(err2)
			assert.Equal(t, err0, appErr.Err)
			assert.Equal(t, err2.Error(), appErr.Message)
			assert.Equal(t, 400, appErr.StatusCode)
		}
	})
}

func TestWithStatusCode(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := WithStatusCode(nil, 200)
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("original")

		err1 := WithStatusCode(err0, 200)

		appErr := Unwrap(err1)
		assert.Equal(t, err0, appErr.Err)
		assert.Equal(t, "", appErr.Message)
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("original")

		err1 := &Error{
			Err:        err0,
			Message:    "message 1",
			StatusCode: 400,
		}
		err2 := WithStatusCode(err1, 500)

		{
			appErr := Unwrap(err1)
			assert.Equal(t, err0, appErr.Err)
			assert.Equal(t, err1.Error(), appErr.Message)
			assert.Equal(t, 400, appErr.StatusCode)
		}

		{
			appErr := Unwrap(err2)
			assert.Equal(t, err0, appErr.Err)
			assert.Equal(t, err1.Error(), appErr.Message)
			assert.Equal(t, 500, appErr.StatusCode)
		}
	})
}

func TestWithReport(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := WithReport(nil)
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("original")

		err1 := WithReport(err0)

		appErr := Unwrap(err1)
		assert.Equal(t, err0, appErr.Err)
		assert.Equal(t, "", appErr.Message)
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("original")

		err1 := WithReport(err0)
		err2 := WithReport(err1)

		{
			appErr := Unwrap(err1)
			assert.Equal(t, err0, appErr.Err)
			assert.Equal(t, true, appErr.Report)
		}

		{
			appErr := Unwrap(err2)
			assert.Equal(t, err0, appErr.Err)
			assert.Equal(t, true, appErr.Report)
		}
	})
}

func TestWrap(t *testing.T) {
	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("original")

		err1 := wrapOrigin(err0)
		assert.Equal(t, "original", err1.Error())

		appErr := Unwrap(err1)
		assert.Equal(t, err0, appErr.Err)
		assert.Equal(t, "", appErr.Message)
		assert.NotEmpty(t, appErr.StackTrace)
		assert.Equal(t, "wrapOrigin", appErr.StackTrace[0].Func)
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("original")

		err1 := wrapOrigin(err0)
		err2 := wrapOrigin(err1)
		assert.Equal(t, "original", err2.Error())

		appErr := Unwrap(err2)
		assert.Equal(t, err0, appErr.Err)
		assert.Equal(t, "", appErr.Message)
		assert.NotEmpty(t, appErr.StackTrace)
		assert.Equal(t, "wrapOrigin", appErr.StackTrace[0].Func)
	})

	t.Run("with pkg/errors", func(t *testing.T) {
		t.Run("pkg/errors.New", func(t *testing.T) {
			err0 := pkgErrorsNew("original")

			err1 := wrapOrigin(err0)
			assert.Equal(t, "original", err1.Error())

			appErr := Unwrap(err1)
			assert.Equal(t, err0, appErr.Err)
			assert.Equal(t, "message: original", appErr.Message)
			assert.NotEmpty(t, appErr.StackTrace)
			assert.Equal(t, "pkgErrorsNew", appErr.StackTrace[0].Func)
		})

		t.Run("pkg/errors.Wrap", func(t *testing.T) {
			err0 := errors.New("original")
			err1 := pkgErrorsWrap(err0, "message")

			err2 := wrapOrigin(err1)
			assert.Equal(t, "original", err2.Error())

			appErr := Unwrap(err2)
			assert.Equal(t, err0, appErr.Err)
			assert.Equal(t, "message: original", appErr.Message)
			assert.NotEmpty(t, appErr.StackTrace)
			assert.Equal(t, "pkgErrorsWrap", appErr.StackTrace[0].Func)
		})
	})
}

func wrapOrigin(err error) error {
	return func() error {
		return wrap(err)
	}()
}