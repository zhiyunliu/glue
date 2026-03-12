package xdb

func GetRowDataReader(obj any) RowDataReader {
	if obj == nil {
		return nil
	}
	reader, ok := obj.(RowDataReader)
	if ok {
		return reader
	}
	return nil
}
