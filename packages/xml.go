// name: xml
// description: XML parsing and manipulation utilities
// author: roturbot
// requires: encoding/xml, encoding/base64, strings

type OSLXML struct {
	document *xmlConfig
	Root     string
}

type xmlConfig struct {
	Root interface{}
}

type XMLNode struct {
	Name      string
	InnerText string
	Attrs     map[string]string
	Children  []XMLNode
}

func NewXML() *OSLXML {
	return &OSLXML{
		document: &xmlConfig{},
	}
}

func (XML) Parse(source any) *OSLXML {
	sourceStr := OSLtoString(source)
	root := XMLNode{
		Name:      "root",
		InnerText: sourceStr,
		Children:  []XMLNode{},
	}

	decoder := xml.NewDecoder(strings.NewReader(sourceStr))

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			attrs := make(map[string]string)
			for _, attr := range se.Attr {
				attrs[attr.Name.Local] = attr.Value
			}
			root.Children = append(root.Children, XMLNode{
				Name:     se.Name.Local,
				Attrs:    attrs,
				Children: []XMLNode{},
			})
		case xml.CharData:
			text := strings.TrimSpace(string(se))
			if len(root.Children) > 0 && text != "" {
				root.Children[len(root.Children)-1].InnerText = text
			}
		}
	}

	return &OSLXML{
		document: &xmlConfig{Root: root},
		Root:     "root",
	}
}

func (XML) Empty() *OSLXML {
	return &OSLXML{
		document: &xmlConfig{},
		Root:     "root",
	}
}

func (x *OSLXML) toStr() string {
	if x.document == nil {
		return ""
	}
	if root, ok := x.document.Root.(XMLNode); ok {
		return x.serializeNode(root)
	}
	return ""
}

func (x *OSLXML) serializeNode(node XMLNode) string {
	var out strings.Builder

	if node.Name == "" {
		return node.InnerText
	}

	out.WriteString("<" + node.Name)

	for k, v := range node.Attrs {
		out.WriteString(fmt.Sprintf(" %s=\"%s\"", k, v))
	}

	if len(node.Children) == 0 && node.InnerText == "" {
		out.WriteString("/>")
	} else {
		out.WriteString(">")
		out.WriteString(node.InnerText)
		for _, child := range node.Children {
			out.WriteString(x.serializeNode(child))
		}
		out.WriteString("</" + node.Name + ">")
	}

	return out.String()
}

func (x *OSLXML) toArr() []any {
	if x.document == nil {
		return []any{}
	}
	if root, ok := x.document.Root.(XMLNode); ok {
		children := make([]any, len(root.Children))
		for i, child := range root.Children {
			children[i] = x.nodeToMap(child)
		}
		return children
	}
	return []any{}
}

func (x *OSLXML) nodeToMap(node XMLNode) map[string]any {
	result := map[string]any{
		"name":  node.Name,
		"text":  node.InnerText,
		"attrs": node.Attrs,
	}

	if len(node.Children) > 0 {
		children := make([]any, len(node.Children))
		for i, child := range node.Children {
			children[i] = x.nodeToMap(child)
		}
		result["children"] = children
	}

	return result
}

func (x *OSLXML) get(path any) map[string]any {
	pathStr := OSLtoString(path)
	parts := strings.Split(pathStr, ">")

	if x.document == nil {
		return nil
	}

	root, ok := x.document.Root.(XMLNode)
	if !ok {
		return nil
	}

	return x.findNode(root, parts, 0)
}

func (x *OSLXML) findNode(node XMLNode, parts []string, index int) map[string]any {
	if index >= len(parts) {
		return x.nodeToMap(node)
	}

	targetName := strings.TrimSpace(parts[index])
	if node.Name == targetName {
		if index == len(parts)-1 {
			return x.nodeToMap(node)
		}

		for _, child := range node.Children {
			if result := x.findNode(child, parts, index+1); result != nil {
				return result
			}
		}
	}

	for _, child := range node.Children {
		if result := x.findNode(child, parts, index); result != nil {
			return result
		}
	}

	return nil
}

func (x *OSLXML) getAll(path any) []any {
	pathStr := OSLtoString(path)
	parts := strings.Split(pathStr, ">")

	if x.document == nil {
		return []any{}
	}

	root, ok := x.document.Root.(XMLNode)
	if !ok {
		return []any{}
	}

	var results []any
	x.findAllNodes(root, parts, 0, &results)
	return results
}

func (x *OSLXML) findAllNodes(node XMLNode, parts []string, index int, results *[]any) {
	if index >= len(parts) {
		*results = append(*results, x.nodeToMap(node))
		return
	}

	targetName := strings.TrimSpace(parts[index])
	if node.Name == targetName {
		if index == len(parts)-1 {
			*results = append(*results, x.nodeToMap(node))
		} else {
			for _, child := range node.Children {
				x.findAllNodes(child, parts, index+1, results)
			}
		}
	} else {
		for _, child := range node.Children {
			x.findAllNodes(child, parts, index, results)
		}
	}
}

func (x *OSLXML) getAttr(path any, attr any) string {
	node := x.get(path)
	if node == nil {
		return ""
	}
	attrs, ok := node["attrs"].(map[string]string)
	if !ok {
		return ""
	}
	return attrs[OSLtoString(attr)]
}

func (x *OSLXML) getChildren(path any) []any {
	node := x.get(path)
	if node == nil {
		return []any{}
	}
	children, ok := node["children"].([]any)
	if !ok {
		return []any{}
	}
	return children
}

func (x *OSLXML) getText(path any) string {
	node := x.get(path)
	if node == nil {
		return ""
	}
	text, ok := node["text"].(string)
	if !ok {
		return ""
	}
	return text
}

func (x *OSLXML) set(path any, tag any) {
	pathStr := OSLtoString(path)
	tagStr := OSLtoString(tag)

	if x.document == nil {
		return
	}

	root, ok := x.document.Root.(XMLNode)
	if !ok {
		return
	}

	tagNode := x.parseTag(tagStr)
	parts := strings.Split(pathStr, ">")

	x.replaceNode(root, parts, 0, tagNode)
}

func (x *OSLXML) parseTag(tagStr string) XMLNode {
	name := ""
	attrs := make(map[string]string)
	innerText := ""

	if strings.HasPrefix(tagStr, "<") {
		endTag := strings.Index(tagStr, ">")
		if endTag == -1 {
			return XMLNode{Name: "", InnerText: tagStr, Attrs: attrs, Children: []XMLNode{}}
		}

		content := tagStr[1:endTag]
		spaceIdx := strings.Index(content, " ")

		if spaceIdx == -1 {
			name = content
		} else {
			name = content[:spaceIdx]
			attrStr := content[spaceIdx+1:]
			attrStr = strings.TrimSpace(attrStr)

			attrStrings := strings.Split(attrStr, " ")
			for _, attrPair := range attrStrings {
				if equalIdx := strings.Index(attrPair, "="); equalIdx > 0 {
					key := attrPair[:equalIdx]
					value := attrPair[equalIdx+1:]
					if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
						value = value[1 : len(value)-1]
					}
					attrs[key] = value
				}
			}
		}

		closingTag := fmt.Sprintf("</%s>", name)
		if closingIdx := strings.Index(tagStr[endTag:], closingTag); closingIdx != -1 {
			innerText = tagStr[endTag+1 : endTag+closingIdx]
		} else {
			innerText = strings.TrimSpace(tagStr[endTag+1:])
		}
	} else {
		innerText = tagStr
	}

	return XMLNode{
		Name:      name,
		InnerText: innerText,
		Attrs:     attrs,
		Children:  []XMLNode{},
	}
}

func (x *OSLXML) replaceNode(node *XMLNode, parts []string, index int, newNode XMLNode) bool {
	if index == len(parts)-1 && node.Name == strings.TrimSpace(parts[index]) {
		*node = newNode
		return true
	}

	targetName := strings.TrimSpace(parts[index])

	if index == len(parts)-1 {
		if node.Name == targetName {
			*node = newNode
			return true
		}
	} else {
		targetNextName := strings.TrimSpace(parts[index+1])
		for i, child := range node.Children {
			if child.Name == targetNextName || (i == 0 && targetName == "*") {
				if x.replaceNode(&node.Children[i], parts, index+1, newNode) {
					return true
				}
			}
		}
	}

	return false
}

func (x *OSLXML) setAttr(path any, attr any, value any) {
	node := x.get(path)
	if node == nil {
		return
	}
	attrs, ok := node["attrs"].(map[string]string)
	if !ok {
		return
	}
	attrs[OSLtoString(attr)] = OSLtoString(value)
}

func (x *OSLXML) setText(path any, value any) {
	pathStr := OSLtoString(path)

	if x.document == nil {
		return
	}

	root, ok := x.document.Root.(XMLNode)
	if !ok {
		return
	}

	parts := strings.Split(pathStr, ">")

	x.updateTextNode(&root, parts, 0, OSLtoString(value))
}

func (x *OSLXML) updateTextNode(node *XMLNode, parts []string, index int, text string) bool {
	if index == len(parts)-1 && node.Name == strings.TrimSpace(parts[index]) {
		node.InnerText = text
		return true
	}

	targetName := strings.TrimSpace(parts[index])

	if index == len(parts)-1 {
		if node.Name == targetName {
			node.InnerText = text
			return true
		}
	} else {
		for i := range node.Children {
			if node.Children[i].Name == strings.TrimSpace(parts[index+1]) || (i == 0 && targetName == "*") {
				if x.updateTextNode(&node.Children[i], parts, index+1, text) {
					return true
				}
			}
		}
	}

	return false
}

func (x *OSLXML) add(path any, tag any) {
	pathStr := OSLtoString(path)
	tagStr := OSLtoString(tag)

	if x.document == nil {
		return
	}

	root, ok := x.document.Root.(XMLNode)
	if !ok {
		return
	}

	tagNode := x.parseTag(tagStr)
	parts := strings.Split(pathStr, ">")

	x.addNodeTo(root, parts, 0, tagNode)
}

func (x *OSLXML) addNodeTo(node XMLNode, parts []string, index int, newNode XMLNode) bool {
	if index == len(parts)-1 && node.Name == strings.TrimSpace(parts[index]) {
		node.Children = append(node.Children, newNode)
		return true
	}

	for i := range node.Children {
		if node.Children[i].Name == strings.TrimSpace(parts[index+1]) || (index == 0 && strings.TrimSpace(parts[index]) == "*") {
			if x.addNodeTo(node.Children[i], parts, index+1, newNode) {
				return true
			}
		}
	}

	return false
}

func (x *OSLXML) remove(path any) {
	pathStr := OSLtoString(path)

	if x.document == nil {
		return
	}

	root, ok := x.document.Root.(XMLNode)
	if !ok {
		return
	}

	parts := strings.Split(pathStr, ">")

	x.removeNode(&root, parts, 0)
}

func (x *OSLXML) removeNode(node *XMLNode, parts []string, index int) bool {
	if index == len(parts)-1 {
		return false
	}

	targetName := strings.TrimSpace(parts[index])
	targetNextName := strings.TrimSpace(parts[index+1])

	if index == len(parts)-2 {
		for i, child := range node.Children {
			if child.Name == targetNextName {
				node.Children = append(node.Children[:i], node.Children[i+1:]...)
				return true
			}
		}
	} else {
		for i := range node.Children {
			if node.Children[i].Name == targetNextName {
				if x.removeNode(&node.Children[i], parts, index+1) {
					return true
				}
			}
		}
	}

	return false
}

func (x *OSLXML) clear(path any) {
	pathStr := OSLtoString(path)

	if x.document == nil {
		return
	}

	root, ok := x.document.Root.(XMLNode)
	if !ok {
		return
	}

	parts := strings.Split(pathStr, ">")

	x.clearNode(&root, parts, 0)
}

func (x *OSLXML) clearNode(node *XMLNode, parts []string, index int) bool {
	if index == len(parts)-1 && node.Name == strings.TrimSpace(parts[index]) {
		node.Children = []XMLNode{}
		node.InnerText = ""
		return true
	}

	targetName := strings.TrimSpace(parts[index])

	for i := range node.Children {
		if node.Children[i].Name == strings.TrimSpace(parts[index+1]) || (index == 0 && targetName == "*") {
			if x.clearNode(&node.Children[i], parts, index+1) {
				return true
			}
		}
	}

	return false
}

func (x *OSLXML) has(path any) bool {
	return x.get(path) != nil
}

func (x *OSLXML) hasAttr(path any, attr any) bool {
	node := x.get(path)
	if node == nil {
		return false
	}
	attrs, ok := node["attrs"].(map[string]string)
	if !ok {
		return false
	}
	attrStr := OSLtoString(attr)
	_, exists := attrs[attrStr]
	return exists
}

func (x *OSLXML) count(path any) int {
	children := x.getChildren(path)
	return len(children)
}

var xml = XML{}
