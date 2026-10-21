package engine

import "testing"

type deepCopyTestItem struct {
	Name string
	Tags []string
}

type deepCopyTestContainer struct {
	Item   *deepCopyTestItem
	Items  []deepCopyTestItem
	Values map[string]*deepCopyTestItem
	Extra  interface{}
	Array  [2]*deepCopyTestItem
}

func TestDeepCopyIsolatesNestedValues(t *testing.T) {
	original := &deepCopyTestContainer{
		Item: &deepCopyTestItem{
			Name: "original",
			Tags: []string{"one", "two"},
		},
		Items: []deepCopyTestItem{{Name: "item", Tags: []string{"item-tag"}}},
		Values: map[string]*deepCopyTestItem{
			"value": {Name: "map-value", Tags: []string{"map-tag"}},
		},
		Extra: &deepCopyTestItem{Name: "interface-value", Tags: []string{"interface-tag"}},
		Array: [2]*deepCopyTestItem{{Name: "array-value", Tags: []string{"array-tag"}}},
	}

	copied := deepCopy(original)
	copied.Item.Name = "changed"
	copied.Item.Tags[0] = "changed"
	copied.Items[0].Tags[0] = "changed"
	copied.Values["value"].Name = "changed"
	copied.Values["value"].Tags[0] = "changed"
	copied.Extra.(*deepCopyTestItem).Tags[0] = "changed"
	copied.Array[0].Name = "changed"

	if original.Item.Name != "original" || original.Item.Tags[0] != "one" {
		t.Fatalf("修改副本的指针字段影响了 original: %+v", original.Item)
	}
	if original.Items[0].Tags[0] != "item-tag" {
		t.Fatalf("修改副本的切片字段影响了 original: %+v", original.Items[0])
	}
	if original.Values["value"].Name != "map-value" || original.Values["value"].Tags[0] != "map-tag" {
		t.Fatalf("修改副本的 map 值影响了 original: %+v", original.Values["value"])
	}
	if original.Extra.(*deepCopyTestItem).Tags[0] != "interface-tag" {
		t.Fatalf("修改副本的 interface 值影响了 original: %+v", original.Extra)
	}
	if original.Array[0].Name != "array-value" {
		t.Fatalf("修改副本的数组元素影响了 original: %+v", original.Array)
	}

	if copied == original || copied.Item == original.Item || &copied.Items[0] == &original.Items[0] {
		t.Fatal("deepCopy 未创建独立对象")
	}
}

func TestDeepCopyPreservesNilValues(t *testing.T) {
	var original *deepCopyTestContainer
	if copied := deepCopy(original); copied != nil {
		t.Fatalf("nil 指针深拷贝结果应为 nil，得到 %+v", copied)
	}

	originalMap := map[string]*deepCopyTestItem(nil)
	copiedMap := deepCopy(originalMap)
	if copiedMap != nil {
		t.Fatalf("nil map 深拷贝结果应为 nil，得到 %+v", copiedMap)
	}

	var originalSlice []deepCopyTestItem
	copiedSlice := deepCopy(originalSlice)
	if copiedSlice != nil {
		t.Fatalf("nil 切片深拷贝结果应为 nil，得到 %+v", copiedSlice)
	}
}
