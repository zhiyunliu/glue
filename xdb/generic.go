package xdb

func GetDataRowReader(obj any) DataRowReader {
	if obj == nil {
		return nil
	}
	reader, ok := obj.(DataRowReader)
	if ok {
		return reader
	}
	return nil
}
