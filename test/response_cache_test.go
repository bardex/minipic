package test

/*
func TestResponseCache(t *testing.T) {
	tests := []struct {
		headers http.Header
		body    []byte
	}{
		{
			headers: http.Header{
				"Content-Type":      {"image/png"},
				"Transfer-Encoding": {"chunked"},
				"Date":              {"Fri, 26 Nov 2021 11:06:48 GMT"},
			},
			body: []byte{32, 33, 34, 35},
		},
		{
			headers: nil,
			body:    []byte{32},
		},
		{
			headers: http.Header{
				"Content-Type":      {"image/png"},
				"Transfer-Encoding": {"chunked"},
				"Date":              {"Fri, 26 Nov 2021 11:06:48 GMT"},
				"X-Test":            {strings.Repeat("s", 1000)},
			},
			body: bytes.Repeat([]byte{32}, 1000000),
		},
	}

	for n, tt := range tests {
		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			fname := fmt.Sprintf("/tmp/%d", time.Nanosecond)
			defer os.Remove(fname)

			rc := app.ResponseCache{}
			rc.filename = fname

			rc.SetHeader(tt.headers)
			rc.Write(tt.body)
			rc.Close()

			rw := httptest.NewRecorder()
			err := rc.WriteTo(rw)
			require.NoError(t, err)
			if tt.headers != nil {
				for k, v := range tt.headers {
					require.Equal(t, v[0], rw.Header().GetItem(k))
				}
			}
			cntLen, err := strconv.Atoi(rw.Header().GetItem("Content-Length"))
			require.Equal(t, len(tt.body), cntLen)
			require.Equal(t, tt.body, rw.Body.Bytes())
			err = rc.Remove()
			require.NoError(t, err)
		})
	}
}
*/
