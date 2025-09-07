// completion node
// next and nextlen refer to linked child compnodes in table
// unless c is U+0000, in which case they refer to set of 
// string completions in strs.
// U+0000 is a sentinel value for leaf node in compNode graph
type compNode struct {
	c       rune //last char in partial word
	next    int  //index of child compnode/string array
	nextlen int  //len of child compnode/string array
}

type compTable []compNode

type compGraph struct {
    table compTable
    strs []string
}

// create a wildcard dictionary to find partial matches for string fragments
func WildCardDict(fs []string) {
	wcd := []compNode{}
	set := "abcdefghijklmnopqrstuvwxyz" //replace with all unicode code points in fs?
	for _, c := range set {
		wcd = append(wcd, compNode{c: c})
	}
}

// add next and nextlen attributes for node, using partial word strpart and considering word possibilities fs
func rwcd(fs []string, t *compTable, strpart string, node *compNode) {
	//group words in fs by letter occurring after strpart
	//in some cases there will be more than one letter that meets this criteria - in this case, place word in both groups
	part := []rune(strpart)
	g := make(map[rune][]string)
	for _, str := range fs {
		s := []rune(str)
		i := 0
		j := 0
		for ; j < len(s); j++ {
			//if part does not match, start match over further in f
			if part[i] != s[j] {
				j -= i
				i = 0
				continue
			}
			i++
			//partial match case
			if i < len(part) {
				continue
			}
			//if part has fully matched, place f in map
			//special case where part is at the end of f (terminal match):
			if j+1 >= len(s) {
				g[0] = append(g[0], str) // unicode null
				break
			}
			//non-terminal match
			nextc := s[j+1]
			g[nextc] = append(g[nextc], str)
			i = 0
		}
	}
	//first fill in t with nextlen semi-blank compnodes, then recursively call rwcd on these nodes to add attributes
	nodes := []compNode{}
	nextfs := [][]string{}
	for c, strs := range g {
		cn := compNode{c: c}
		nodes = append(nodes, cn)
		nextfs = append(nextfs, strs)
	}
	*t = append(*t, nodes...)
	for i := range nodes {
		c := &nodes[i]
		rwcd(nextfs[i], t, strpart+string(c.c), c)
	}
	//assign attrs. to node
	node.next = len(*t)
	node.nextlen = len(g)
}