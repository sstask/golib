package stconfig

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"

	xmlx "github.com/jteeuwen/go-pkg-xmlx"
)

type XmlNode struct {
	Inner *xmlx.Node
}

func LoadXml(file string) (*XmlNode, error) {
	doc := xmlx.New()
	if err := doc.LoadFile(file, nil); err != nil {
		return nil, err
	}
	return &XmlNode{doc.Root}, nil
}

func (this *XmlNode) SaveXml(path string) error {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="utf-8" ?>`)
	b.WriteByte('\n')
	b.Write(this.Inner.Bytes())
	return ioutil.WriteFile(path, b.Bytes(), 0600)
}

//nodename can be splited by '/'
//for example "a/b", it find nodes like "<a><b></b></a>" but not "<a><c></c><b></b></a>"
func (curnode *XmlNode) FindNode(nodename string) *XmlNode {
	ret := &XmlNode{nil}
	if curnode.Inner == nil {
		return ret
	}

	names := strings.Split(nodename, "/")
	var node *xmlx.Node
	for _, v := range names {
		if node == nil {
			node = curnode.Inner.SelectNode("*", v)
			if node == nil {
				return ret
			}
		} else {
			nodes := node.SelectNodesDirect("*", v)
			if len(nodes) == 0 {
				return ret
			} else {
				node = nodes[0]
			}
		}
	}
	if node == nil {
		return ret
	}
	return &XmlNode{node}
}

func (curnode *XmlNode) FindNodes(nodename string) []*XmlNode {
	if curnode.Inner == nil {
		return nil
	}
	names := strings.Split(nodename, "/")
	var node *xmlx.Node
	for i := 0; i < len(names)-1; i++ {
		if node == nil {
			node = curnode.Inner.SelectNode("*", names[i])
			if node == nil {
				return nil
			}
		} else {
			nodes := node.SelectNodesDirect("*", names[i])
			if len(nodes) == 0 {
				return nil
			} else {
				node = nodes[0]
			}
		}
	}
	var nodes []*xmlx.Node
	if node == nil {
		nodes = curnode.Inner.SelectNodes("", nodename)
	} else {
		nodes = node.SelectNodes("", nodename)
	}
	if len(nodes) > 0 {
		res := make([]*XmlNode, 0, len(nodes))
		for _, v := range nodes {
			res = append(res, &XmlNode{v})
		}
		return res
	}

	return nil
}

func findnodebyattr(node *xmlx.Node, attr string, val string) *xmlx.Node {
	if node == nil {
		return nil
	}
	if node.HasAttr("*", attr) && node.As("*", attr) == val {
		return node
	}
	for _, v := range node.Children {
		if v.HasAttr("*", attr) && v.As("*", attr) == val {
			return v
		} else {
			res := findnodebyattr(v, attr, val)
			if res != nil && res.HasAttr("*", attr) && res.As("*", attr) == val {
				return res
			}
		}
	}
	return nil
}

func (node *XmlNode) FindNodeByAttr(attr string, val string) *XmlNode {
	return &XmlNode{findnodebyattr(node.Inner, attr, val)}
}

func (node *XmlNode) GetVal() string {
	if node.Inner == nil {
		return ""
	}
	return node.Inner.GetValue()
}

func (node *XmlNode) GetAttr(attrname string) string {
	if node.Inner == nil {
		return ""
	}
	return node.Inner.As("", attrname)
}

func (node *XmlNode) GetValI() int {
	if node.Inner == nil {
		return 0
	}
	value := node.Inner.GetValue()
	if value != "" {
		n, _ := strconv.ParseInt(value, 10, 0)
		return int(n)
	}
	return 0
}

func (node *XmlNode) GetAttrI(attrname string) int {
	if node.Inner == nil {
		return 0
	}
	return node.Inner.Ai("", attrname)
}

func (node *XmlNode) SetVal(val string) {
	if node.Inner != nil {
		node.Inner.SetValue(val)
	}
}

func (node *XmlNode) SetAttr(attrname string, val string) {
	if node.Inner != nil {
		node.Inner.SetAttr(attrname, val)
	}
}
