stconfig is a simple lib for reading some config.
example
```
	strs, _ := ReadXlsx("xxx.xlsx", "sss")

	node, _ := LoadXml("test.xml")
	fmt.Println(node.FindNode("title").GetValI())
	node.FindNodeByAttr("name", "sss").SetVal("3")
	node.SaveXml("test1.xml")
```
