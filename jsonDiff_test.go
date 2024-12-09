package colorisediff

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter"
)

func removeANSIColorCodes(input string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(input, "")
}

func TestSprintJSONDiff(t *testing.T) {
	tests := []struct {
		name            string
		json1           string
		json2           string
		expectedStringA []string
		expectedStringB []string
		isNoised        bool
		noise           map[string][]string
	}{
		{
			expectedStringA: []string{
				"b400a1d1cb03c0080f4cf1ce133de291cd61f27bd93764d9a0f8316e65cc745e",
				"0aa5c8db7f9bf982e95f31d774160b25f7c0ef3f1a1b04f5b52401266ede3669",
			},
			expectedStringB: []string{
				"c9e29f82c1f21d8d25ab2fad7e1c87c5223c2598d7b29ca3e33fdfeb18727833",
				"8698e0dee7a9c61f377e96198cd3b735af70070648ac6fbdbc924a9c684b21d4",
			},
			json1: "{\"level1\":{\"level2\":{\"name\":\"Cat\",\"id\":3}}}",
			json2: "{\"level1\":{\"level2\":{\"name\":\"Dog\",\"id\":3}}}",
			name:  "nested JSONs with different structures",
		},
		{
			expectedStringA: []string{
				"3bcc749b84f87efc5fd06c5b77ea853b6fff0c4f3f317f60fb41480096d64597",
			},
			expectedStringB: []string{
				"935be748ebb92097cc80dd5c3b55282b718bb27bcfeff389b5b096fd7165c646",
			},
			json1: "{\"animals\":[{\"name\":\"Cat\"},{\"name\":\"Dog\"},{\"name\":\"Elephant\"}]}",
			json2: "{\"animals\":[{\"name\":\"Dog\"},{\"name\":\"Cat\"},{\"apple\":\"lusiancs\"},{\"name\":\"Elephant\"}]}",
			name:  "nested JSONs with array length differences",
		},
		{
			expectedStringA: []string{
				"3bcc749b84f87efc5fd06c5b77ea853b6fff0c4f3f317f60fb41480096d64597",
			},
			expectedStringB: []string{"6965acb9dc1e8ce2c7a581ebeb35c5f19e1da8b7aa71006bfcc8105509bcfdf3"},
			json1:           "{\"animals\":[{\"name\":\"Cat\"},{\"name\":\"Dog\"},{\"name\":\"Elephant\"}]}",
			json2:           "{\"animals\":[{\"name\":\"Dog\"},{\"name\":\"Cat\"},{\"apple\":\"lusiancs\"}]}",
			name:            "nested JSONs with array differences",
		},
		{
			expectedStringA: []string{
				"025ed6cf861466ccc5332c1fab04824e93ec2cb230757a1ab0904043ae6be5ba", "0584a507040556881693562961f05aba1feb0428ca027e5ee578e4c50b72ec8c", "b47122445819617c5b04cbcfaaa970fcc292b0fb629dee0cf8a11071dd4766a5", "c4d3a315342e20149eb091f28553c84231d9eb56dead99219782202e3b16e036",
			},
			expectedStringB: []string{
				"ff6db42fe1d08a641ffebeba7755129e5abcbb253e9d30e9a3bf6ebf3d6a0df5", "bcc5632a6614fda69d7e7fa9133e5a352c7e004a543a45a3141b38c2abfcead1",
				"81a7fb57f3b57cba03a55c613e20fe8d649ddd9bc64a700cb5284e2b81cdca2c",
				"e40d280fe6a2c2ab7d7f6809b79c45ce62b74f816d00487c81e975ca079bcd01",
			},
			json1: "{\"animal\":{\"name\":\"Cat\",\"attributes\":{\"color\":\"black\",\"age\":5}}}",
			json2: "{\"animal\":{\"name\":\"Cat\",\"attributes\":{\"color\":\"white\",\"age\":5}}}",
			name:  "different key-value pairs in nested JSON",
		},
		{
			expectedStringA: []string{
				"71177a8bc7e1abfe019a4a1fad9407dc547362a59d70d16761ecc8f50d9ab31e",
				"83058f9c21b01a272805827b35b527121b6c3b9a8189e3bded0fc269049c8121",
			},
			expectedStringB: []string{
				"4f0abd2d3a443c3d88da6e2ca4da35556cadce429f7f2f8506ec4a94a99c48d7",
				"3945b6ce91aa49db148fdf52f018def9284a9f01a956abfe5fb1d4b24de634dc",
			},
			json1: "{\"animals\":{\"domestic\":[\"Cat\",\"Dog\"],\"wild\":[\"Elephant\",\"Lion\"]}}",
			json2: "{\"animals\":{\"domestic\":[\"Dog\",\"Cat\"],\"wild\":[\"Lion\",\"Elephant\"]}}",
			name:  "nested arrays within objects",
		},
		{
			expectedStringA: []string{
				"748a984d11fa3e28ac12f17280043926cda016f4ed91b64c1fdfc378dc39840f",
				"ca9366289c771419820364f7f72ebb1f032e4501c9f637b267b4c688e929ba7c",
			},
			expectedStringB: []string{
				"62d5ed8cda82e65e27831497216f5690e3ec428108932e6c50df54d9d38ad9b5",
				"fa135fcd9ff45b6a832ab2d3bdd8d340bd1eac2bea8ba9b6fa2465df7c1d0f71",
			},
			json1: "{\"level1\":{\"level2\":{\"level3\":{\"name\":\"Cat\",\"id\":3}}}}",
			json2: "{\"level1\":{\"level2\":{\"level3\":{\"name\":\"Dog\",\"id\":3}}}}",
			name:  "three layered nested maps",
		},
		{
			expectedStringA: []string{
				"e40b10192f58e3cc848e216beb7a9b4ad2cef79cb042e5a82310eac0cca41d56",
				"5b39bfb46f15e4cbb84fc05e045889c7544974a52654d1c0166fbdd69de6cca5",
				"fd5cb7711aba0fb67c29cc3ae2bfc698b86087522366aef2b781a0001e4191b8",
				"6b4719e9d04cf8251fc223e197c3416a3fb8b2b933a3ad448b8a81862290c21b",
			},
			expectedStringB: []string{"bb43e7d4888ae02aa48c05914962d12a5dd39ef435bb3bea7268699eaf8ce7e0", "6f20a43ca4c698864e7cf4e44f68285232f36f4942116e965095fb77f51878dd"},
			json1:           "{\"animal\":{\"name\":\"Cat\",\"features\":{\"furly\":\"short\",\"tail\":\"long\"}}}",
			json2:           "{\"animal\":{\"name\":\"Cat\",\"features\":{\"fur\":\"long\",\"tail\":\"long\"}}}",
			name:            "nested objects with different keys",
		},
		{
			expectedStringA: []string{
				"30aef022f108468def1ab4e9c2f7bc955dd88234b7fa6d6b4fa36584b9373830",
				"4e749ff18c8ee678247126d751bb1a6309a22e5e754c8b1d4cac0c7701cd96c0",
				"7470cc614de3e68eb00d38f087b208d73822244dca550c3103caa61d7bc16f3f",
				"7fb2c228f44a66eb262804bd7c21049dcfbb265173cf24deb4842e521ee983b2",
				"93639a34e7bd6dd08cc716ed1715bfd6207cc5f8768525dad30564c6a843f32a",
				"d90a8ef41e08f7d80ccee8b2c69d43291933a7eeb57f517469f3a7f14cc1a313",
				"3169f37d3775226c0cf126fc1a15e413b3f83d2c27a80a9a82d91d2761ce03e1",
				"f3b21448b50b0b23a94aed8a44376b958931675c8b216a1ce3929710311e1689",
				"c3e49694ce1509a2bf9f6557c4fc643d1efc1ef98dd9a7c18cbe2fb96bdc5109",
			},
			expectedStringB: []string{
				"fb916cbcb8a9c8accb60f436b7ae49d404a6815d81cb67aae9236d41713ecb2c",
				"8211a861b8ec0d7ce838850e29a9a31a352abcccdb045cb86cd8c8d8721dcd20",
				"ca4aa97a9cd8a928eb843ba02d7a1b8eac3014ed92e7d934fa996c6cf829aa05",
				"32ed13ed95fd695d63687876616d3a2e395137ef5c09d96cff18458eec33e8f3",
				"ef79067ed4dfbf2f0dbe53cb81e5882b6bfa756495dad1969b58d27fa2d170ab",
				"56d4dd69bc7d542a099e6200f2b6d5d024f747e8fc8f493ca9d0a449cc63d1c3",
				"70e5515e928a51a26c2d78315c677bd841b6118ca4f8f9843b349257a0fafa1e",
				"b54c3ce37beedd7f27b98563770c930887474a00f8471252a95a4fd8e4b8b1fa",
				"d458672f27aa045490d25f180b1e5b81d8d6da09035731127ffb27eca9002942",
				"c3e49694ce1509a2bf9f6557c4fc643d1efc1ef98dd9a7c18cbe2fb96bdc5109",
			},
			json1: "{\"zoo\":{\"animals\":[{\"type\":\"mammal\",\"name\":\"Elephant\",\"age\":10},{\"type\":\"bird\",\"name\":\"Parrot\",\"age\":2}]}}",
			json2: "{\"zoo\":{\"animals\":[{\"type\":\"mammal\",\"name\":\"Elephant\",\"age\":10},{\"type\":\"bird\",\"name\":\"Parrot\",\"age\":3}]}}",
			name:  "deeply nested mixed structures",
		},
		{
			expectedStringA: []string{
				"ef4dd42bb9dc629bbfade57743e72663c9a5d236c0e92cb4ce8c80e0d1304350",
				"f71cbcb39804b502c41125d03be6c699baafa15efba8bbb5aef0b111039b2a87",
				"fa012aa280e519ccb558663c5728faea94ed020988c83653a796926c35902ec0",
				"ef0a5b31ffc0a36df02dcc08898cad0b92857cd1405cad0feefc18d888bf57d0",
				"e0236118ff8532288842ad67be5bca9f81b15191ee2efc2eee077406fabf8bbd",
				"c828a0590dbd6e6eefbee21c5855b19a8bff98930f0219816fe6f24c3705c5cb",
				"17f1dc518cda544ae5ff4b2479e94d5ef811e542b387608dfe7b44e42937e452",
			},
			expectedStringB: []string{
				"d79b35acf01b0f5138699ff1cc49ea89373b8ebf7e96118b839586a28c28bbee",
				"8fe4e8830eb84cdacd2cbd60f62fc5d50dcecf3a5cc439ea7e24d87d4257c6a8",
				"e765848380611cb81996ea9908ade2ee8940c21d72a84fd19ce1d1d6ddfa8e2a",
				"001ff4d6bf9821bb067c73812ba5900574dd161d813f10623ba2515fdbed0f88",
				"19018c74ffe402eb59202aadc1cab4f5c8171c96ba50f4621ab9d72f3b18914e",
				"0bd78116662eba5d4fa8bbc64f81afbd879fb2e73cd6d85105e1f9bf3a658ae0",
				"d11e6a5e5047d70f5e5633650bc4b3fd7a588d126fb785e5ac42b92fbf3e44f6",
			},
			json1: "{\"books\":[{\"title\":\"Book A\",\"author\":{\"name\":\"Author 1\"}},{\"title\":\"Book B\",\"author\":{\"name\":\"Author 2\"}}]}",
			json2: "{\"books\":[{\"title\":\"Book B\",\"author\":{\"name\":\"Author 2\"}},{\"title\":\"Book A\",\"author\":{\"name\":\"Author 1\"}}]}",
			name:  "different arrays with nested objects",
		},
		{
			expectedStringA: []string{
				"83a2bdc32cf3b2ecba06bf4ef3c4bb11d6219a8ade62a1026667616221f4343b",
				"a717fd9c62ccea2df24835e9db8cb99f8af3ebddc352a758dc55750e272f5a2b",
			},
			expectedStringB: []string{
				"847155ec0ab7092a7ed5a91b01073f736a03c7e5c0b2c61ea9daeb89fdb680ef",
				"4baa4c73af6838aabe1796bce3a0ea547d4d5efa4bd630cbeeafd82003b0163b",
			},
			json1: "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value1\"}",
			json2: "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value2\"}",
			name:  "map containing array and string",
		},
		{
			expectedStringA: []string{
				"faa2ae50e03a03bd9b555d1422c9a62da0d8ab64bcaafa95d7da7a7f92ee0f5b",
			},
			expectedStringB: []string{
				"dea2a1acabc4ce03b37f536b24a7053f432d5392492f0d9d00f5fe6b0be3909a",
			},
			json1: "{\"level1\": {\"level2\": {\"value\": 10}}}",
			json2: "{\"level1\": {\"level2\": {\"value\": 20}}}",
			name:  "nested maps with number differences",
		},
		{
			expectedStringA: []string{
				"e09f8de90ad75017085449da9978780c5aed1d5148c699bbbbfafe1feb05d2e1",
			},
			expectedStringB: []string{
				"9512ac00c487192872a36662b05bbf16d2c500206a9fc02c3f75b4ce5ab1f195",
			},
			json1: "{\"a\":[{\"b\":[{\"c\":\"d\"},2,3,{\"e\":\"f\"}]},[\"g\",\"h\"]]}",
			json2: "{\"a\":[{\"b\":[{\"c\":\"d\"},3,2,{\"e\":\"f\"}]},[\"h\",\"g\"]]}",
			name:  "complex multi-type nested structures",
		},
		{
			expectedStringA: []string{
				"3cf5725c92bc27fa26481d4c5a686beef5a240f963e42825b3b3da47925b2f99",
			},
			expectedStringB: []string{
				"255666fae88b8a55bb0e8e577e5dd79bc03cbe868bdeda012995963d96928f39",
			},
			json1: `{"nested":{"key":[]}}`,                                          // Empty array.
			json2: `{"nested":{"key":[{"mapKey1":"value1"},{"mapKey2":"value2"}]}}`, // Array of maps.
			name:  "empty array to array of maps",
		},
		{
			expectedStringA: []string{
				"3cf5725c92bc27fa26481d4c5a686beef5a240f963e42825b3b3da47925b2f99",
			},
			expectedStringB: []string{
				"d03634df9355fa94c2dba26bbe8b3acafb5e054e3e2eea579f952756feec18be",
			},
			json1: "{\"nested\":{\"key\":[]}}",
			json2: "{\"nested\":{\"key\":[{\"mapKey1\":\"value1\", \"mapKey2\":[1, 2, {\"subKey\":\"subValue\"}], \"mapKey3\":{\"innerKey\":\"innerValue\"}}, {\"mapKey4\":\"value2\", \"mapKey5\":[3, 4, {\"subKey2\":\"subValue3\"}], \"mapKey6\":{\"innerKey2\":\"innerValue2\"}}]}}",
			name:  "array to complex array of maps",
		},
		{
			expectedStringA: []string{
				"3bcc749b84f87efc5fd06c5b77ea853b6fff0c4f3f317f60fb41480096d64597",
			},
			expectedStringB: []string{
				"935be748ebb92097cc80dd5c3b55282b718bb27bcfeff389b5b096fd7165c646",
			},
			json1: "{\"animals\":[{\"name\":\"Cat\"},{\"name\":\"Dog\"},{\"name\":\"Elephant\"}]}",
			json2: "{\"animals\":[{\"name\":\"Dog\"},{\"name\":\"Cat\"},{\"apple\":\"lusiancs\"},{\"name\":\"Elephant\"}]}",
			name:  "nested JSONs with array length differences",
		},
		{
			expectedStringA: []string{
				"3bcc749b84f87efc5fd06c5b77ea853b6fff0c4f3f317f60fb41480096d64597",
			},
			expectedStringB: []string{
				"6965acb9dc1e8ce2c7a581ebeb35c5f19e1da8b7aa71006bfcc8105509bcfdf3",
			},
			json1: "{\"animals\":[{\"name\":\"Cat\"},{\"name\":\"Dog\"},{\"name\":\"Elephant\"}]}",
			json2: "{\"animals\":[{\"name\":\"Dog\"},{\"name\":\"Cat\"},{\"apple\":\"lusiancs\"}]}",
			name:  "nested JSONs with array differences",
		},
		{
			expectedStringA: []string{
				"b47122445819617c5b04cbcfaaa970fcc292b0fb629dee0cf8a11071dd4766a5",
				"0584a507040556881693562961f05aba1feb0428ca027e5ee578e4c50b72ec8c",
				"025ed6cf861466ccc5332c1fab04824e93ec2cb230757a1ab0904043ae6be5ba",
				"c4d3a315342e20149eb091f28553c84231d9eb56dead99219782202e3b16e036",
			},
			expectedStringB: []string{
				"ff6db42fe1d08a641ffebeba7755129e5abcbb253e9d30e9a3bf6ebf3d6a0df5",
				"81a7fb57f3b57cba03a55c613e20fe8d649ddd9bc64a700cb5284e2b81cdca2c",
				"e40d280fe6a2c2ab7d7f6809b79c45ce62b74f816d00487c81e975ca079bcd01",
				"bcc5632a6614fda69d7e7fa9133e5a352c7e004a543a45a3141b38c2abfcead1",
			},
			json1: "{\"animal\":{\"name\":\"Cat\",\"attributes\":{\"color\":\"black\",\"age\":5}}}",
			json2: "{\"animal\":{\"name\":\"Cat\",\"attributes\":{\"color\":\"white\",\"age\":5}}}",
			name:  "different key-value pairs in nested JSON",
		},
		{
			expectedStringA: []string{
				"748a984d11fa3e28ac12f17280043926cda016f4ed91b64c1fdfc378dc39840f",
				"ca9366289c771419820364f7f72ebb1f032e4501c9f637b267b4c688e929ba7c",
			},
			expectedStringB: []string{
				"62d5ed8cda82e65e27831497216f5690e3ec428108932e6c50df54d9d38ad9b5",
				"fa135fcd9ff45b6a832ab2d3bdd8d340bd1eac2bea8ba9b6fa2465df7c1d0f71",
			},
			json1: "{\"level1\":{\"level2\":{\"level3\":{\"name\":\"Cat\",\"id\":3}}}}",
			json2: "{\"level1\":{\"level2\":{\"level3\":{\"name\":\"Dog\",\"id\":3}}}}",
			name:  "three layered nested maps",
		},
		{
			expectedStringA: []string{
				"5b39bfb46f15e4cbb84fc05e045889c7544974a52654d1c0166fbdd69de6cca5",
				"e40b10192f58e3cc848e216beb7a9b4ad2cef79cb042e5a82310eac0cca41d56",
				"6b4719e9d04cf8251fc223e197c3416a3fb8b2b933a3ad448b8a81862290c21b",
				"fd5cb7711aba0fb67c29cc3ae2bfc698b86087522366aef2b781a0001e4191b8",
			},
			expectedStringB: []string{
				"bb43e7d4888ae02aa48c05914962d12a5dd39ef435bb3bea7268699eaf8ce7e0",
				"6f20a43ca4c698864e7cf4e44f68285232f36f4942116e965095fb77f51878dd",
			},
			json1: "{\"animal\":{\"name\":\"Cat\",\"features\":{\"furly\":\"short\",\"tail\":\"long\"}}}",
			json2: "{\"animal\":{\"name\":\"Cat\",\"features\":{\"fur\":\"long\",\"tail\":\"long\"}}}",
			name:  "nested objects with different keys",
		},
		{
			expectedStringA: []string{
				"7fb2c228f44a66eb262804bd7c21049dcfbb265173cf24deb4842e521ee983b2",
				"30aef022f108468def1ab4e9c2f7bc955dd88234b7fa6d6b4fa36584b9373830",
				"c3e49694ce1509a2bf9f6557c4fc643d1efc1ef98dd9a7c18cbe2fb96bdc5109",
				"7470cc614de3e68eb00d38f087b208d73822244dca550c3103caa61d7bc16f3f",
				"4e749ff18c8ee678247126d751bb1a6309a22e5e754c8b1d4cac0c7701cd96c0",
				"d90a8ef41e08f7d80ccee8b2c69d43291933a7eeb57f517469f3a7f14cc1a313",
				"3169f37d3775226c0cf126fc1a15e413b3f83d2c27a80a9a82d91d2761ce03e1",
				"93639a34e7bd6dd08cc716ed1715bfd6207cc5f8768525dad30564c6a843f32a",
				"f3b21448b50b0b23a94aed8a44376b958931675c8b216a1ce3929710311e1689",
				"c3e49694ce1509a2bf9f6557c4fc643d1efc1ef98dd9a7c18cbe2fb96bdc5109",
				"c3e49694ce1509a2bf9f6557c4fc643d1efc1ef98dd9a7c18cbe2fb96bdc5109",
			},
			expectedStringB: []string{
				"fb916cbcb8a9c8accb60f436b7ae49d404a6815d81cb67aae9236d41713ecb2c",
				"56d4dd69bc7d542a099e6200f2b6d5d024f747e8fc8f493ca9d0a449cc63d1c3",
				"d458672f27aa045490d25f180b1e5b81d8d6da09035731127ffb27eca9002942",
				"8211a861b8ec0d7ce838850e29a9a31a352abcccdb045cb86cd8c8d8721dcd20",
				"b54c3ce37beedd7f27b98563770c930887474a00f8471252a95a4fd8e4b8b1fa",
				"70e5515e928a51a26c2d78315c677bd841b6118ca4f8f9843b349257a0fafa1e",
				"ef79067ed4dfbf2f0dbe53cb81e5882b6bfa756495dad1969b58d27fa2d170ab",
				"ca4aa97a9cd8a928eb843ba02d7a1b8eac3014ed92e7d934fa996c6cf829aa05",
				"32ed13ed95fd695d63687876616d3a2e395137ef5c09d96cff18458eec33e8f3",
			},
			json1: "{\"zoo\":{\"animals\":[{\"type\":\"mammal\",\"name\":\"Elephant\",\"age\":10},{\"type\":\"bird\",\"name\":\"Parrot\",\"age\":2}]}}",
			json2: "{\"zoo\":{\"animals\":[{\"type\":\"mammal\",\"name\":\"Elephant\",\"age\":10},{\"type\":\"bird\",\"name\":\"Parrot\",\"age\":3}]}}",
			name:  "deeply nested mixed structures",
		},
		{
			expectedStringA: []string{
				"d896342371be7600a1d266ad71ce895ddfb3bb307928a2fb7338d5bfb12d16a7",
				"8317e30c64967ec9a7d82332831630315a64ab03496dc11851984693ccc6ef3c",
				"507d00b6db2ba429d7bb6fd7cf422e145806761748690714ae4bd7672bb2e41a",
				"e61a886b811b80831b9a69293200addfefea3ac0a4fbc401422c32b0815353f7",
				"689295d50d283dfb54cef114968a0b5eebd44613ae313bc9d1719e8a18e4c83d",
				"2622053f90dead9c5bbad099944c3736877f2d65ce9bb716b782b2467936b27c",
				"e94990f2032dca084904454bdd78da23ae7de904aea3620b4b6e3a0b0ff85f96",
				"52c0d440ad1a7658e150f9fd60328042f32cff5695290d00af86b915f814d55b",
				"7b34250150643662fc5e023ab8793533c3929acbe80f5202a3177d2c37e5fe79",
				"76ff030bf9645182aacf9383e97f78ec96143e7d79c56a8fe4b28440f4fc6092",
				"3b477686293f714247a3f70386610e2dfeec8ec161eaee76d50025433579955f",
				"e3a93f1273fe682c72e5e440f9e706d97765baaafb468809072759f20c024dca",
				"16c0692326655ceedeaccadaa05f04fed9c39a9b8b25783c0401cec429de5e6d",
				"897eb927925bc216402ad7206e3a8ac49c834e3958900786667bd714ef1b8f1f",
				"b3f335d396b18a08d5045a9af8bc9319205067a3f0242139093295c151d26d38",
				"ed177368aab7cdc3b865a1b211491689beb12ac9a2092bc302aa8e9b207de37b",
			},
			expectedStringB: []string{
				"82ef5e26330856df883b55e13725a1b71c56f787ab8b8e1d47d6df69ddf7121d",
				"275e4bec238e61b624affaf51ece1beaa5e1d3cb08d875a501e6f0a32e6b6474",
				"5b43d1d31702c9e74b93420fe30a3e64988edeb18b5aa7c6030108ab5ca43ccc",
				"7112d7ed41c984911e7716bf7791972e4d4712e59578a9a6efe565632c162076",
				"cf358de4037ee1daf5a36d4b6b89ded7aced41b6ac1a8ec0bcc9e063607eb194",
				"78970da2cca7e3dfa6c46b4136603259b56eccc3083d51cd2bc190648e7f9dc1",
				"24b7638a8f6fe9e626945f3cd9d85d166a47c98b502dc6e48a91ec254cb857c2",
				"bf7120f76e1a3722fc44a6a21b2de827648b5c30a337d13a7037c3c51ee0615f",
				"27fc604f00be8e93b2e8ae0d8bc44c371c5c71b172c8cd7c83c1bed5cba0329d",
				"04323636cbf90c4e5faed7090d42c598792e3abec38503e07affaf9b476dd2e2",
				"a7e87a1d0a3ca5f8a1c79e576883489884cc72f607495418f6c7910cc6945941",
			},
			json1: "{\"family\":{\"parents\":[{\"name\":\"Alice\",\"age\":40},{\"name\":\"Bob\",\"age\":42}],\"children\":[{\"name\":\"Charlie\",\"age\":10},{\"name\":\"Daisy\",\"age\":8}]}}",
			json2: "{\"family\":{\"parents\":[{\"name\":\"Bob\",\"age\":42},{\"name\":\"Alice\",\"age\":40}],\"children\":[{\"name\":\"Daisy\",\"age\":8},{\"name\":\"Charlie\",\"age\":10}]}}",
			name:  "complex nested objects and arrays",
		},
		{
			expectedStringA: []string{
				"fa012aa280e519ccb558663c5728faea94ed020988c83653a796926c35902ec0",
				"ef4dd42bb9dc629bbfade57743e72663c9a5d236c0e92cb4ce8c80e0d1304350",
				"ef0a5b31ffc0a36df02dcc08898cad0b92857cd1405cad0feefc18d888bf57d0",
				"f71cbcb39804b502c41125d03be6c699baafa15efba8bbb5aef0b111039b2a87",
				"e0236118ff8532288842ad67be5bca9f81b15191ee2efc2eee077406fabf8bbd",
				"17f1dc518cda544ae5ff4b2479e94d5ef811e542b387608dfe7b44e42937e452",
			},
			expectedStringB: []string{
				"8fe4e8830eb84cdacd2cbd60f62fc5d50dcecf3a5cc439ea7e24d87d4257c6a8",
				"e765848380611cb81996ea9908ade2ee8940c21d72a84fd19ce1d1d6ddfa8e2a",
				"d79b35acf01b0f5138699ff1cc49ea89373b8ebf7e96118b839586a28c28bbee",
				"001ff4d6bf9821bb067c73812ba5900574dd161d813f10623ba2515fdbed0f88",
				"19018c74ffe402eb59202aadc1cab4f5c8171c96ba50f4621ab9d72f3b18914e",
				"d11e6a5e5047d70f5e5633650bc4b3fd7a588d126fb785e5ac42b92fbf3e44f6",
			},
			json1: "{\"books\":[{\"title\":\"Book A\",\"author\":{\"name\":\"Author 1\"}},{\"title\":\"Book B\",\"author\":{\"name\":\"Author 2\"}}]}",
			json2: "{\"books\":[{\"title\":\"Book B\",\"author\":{\"name\":\"Author 2\"}},{\"title\":\"Book A\",\"author\":{\"name\":\"Author 1\"}}]}",
			name:  "different arrays with nested objects",
		},
		{
			expectedStringA: []string{
				"a46a95231b9cd0ad10bc0cfbaa15b106a626f028026f5fb1c04c277706dea4ba",
				"8497a6d6bc8e7badf16c32683f73374b9381fb6d28c0baa3371c7bffbe363116",
				"a46a95231b9cd0ad10bc0cfbaa15b106a626f028026f5fb1c04c277706dea4ba",
				"8497a6d6bc8e7badf16c32683f73374b9381fb6d28c0baa3371c7bffbe363116",
			},
			expectedStringB: []string{
				"30490381e549259178bb23d3c2ff563c0ab85843b9408a547f6de169510e9e27",
				"aa6dc46abfebfb022c2223c1e58ab204224a3e5bb20189c677c7ec1b13dd67e8",
				"30490381e549259178bb23d3c2ff563c0ab85843b9408a547f6de169510e9e27",
				"aa6dc46abfebfb022c2223c1e58ab204224a3e5bb20189c677c7ec1b13dd67e8",
			},
			json1: "{\"outer\": {\"inner\": [{\"key\": \"value1\"}, {\"key\": \"value2\"}], \"array\": [1, 2, 3]}}",
			json2: "{\"outer\": {\"inner\": [{\"key\": \"value1\"}, {\"key\": \"value3\"}], \"array\": [1, 3, 2]}}",
			name:  "complex nested structures with maps and arrays",
		},
		{
			expectedStringA: []string{
				"faa2ae50e03a03bd9b555d1422c9a62da0d8ab64bcaafa95d7da7a7f92ee0f5b",
			},
			expectedStringB: []string{
				"dea2a1acabc4ce03b37f536b24a7053f432d5392492f0d9d00f5fe6b0be3909a",
			},
			json1: "{\"level1\": {\"level2\": {\"value\": 10}}}",
			json2: "{\"level1\": {\"level2\": {\"value\": 20}}}",
			name:  "nested maps with number differences",
		},
		{
			expectedStringA: []string{
				"3cf5725c92bc27fa26481d4c5a686beef5a240f963e42825b3b3da47925b2f99",
			},
			expectedStringB: []string{
				"255666fae88b8a55bb0e8e577e5dd79bc03cbe868bdeda012995963d96928f39",
			},
			json1: `{"nested":{"key":[]}}`, // Empty array.
			json2: `{"nested":{"key":[{"mapKey1":"value1"},{"mapKey2":"value2"}]}}`,
			name:  "empty array to array of maps",
		},
		{
			expectedStringA: []string{
				"3cf5725c92bc27fa26481d4c5a686beef5a240f963e42825b3b3da47925b2f99",
			},
			expectedStringB: []string{
				"d03634df9355fa94c2dba26bbe8b3acafb5e054e3e2eea579f952756feec18be",
			},
			json1: "{\"nested\":{\"key\":[]}}",
			json2: "{\"nested\":{\"key\":[{\"mapKey1\":\"value1\", \"mapKey2\":[1, 2, {\"subKey\":\"subValue\"}], \"mapKey3\":{\"innerKey\":\"innerValue\"}}, {\"mapKey4\":\"value2\", \"mapKey5\":[3, 4, {\"subKey2\":\"subValue3\"}], \"mapKey6\":{\"innerKey2\":\"innerValue2\"}}]}}",
			name:  "empty array to complex array of maps",
		},
		{
			expectedStringA: []string{
				"49bec237abb42a872e82edee006cb72e8270b4c14179140ce03ebc47ad36fa2d",
				"3cd84203cf23bffc56c12344c2b2fbf313c1e4ed34f125ea813a50b42adca1d9",
				"825c067292a1b5ce6ce1724c52fa2068bfb651a35273a65351be3e63d8614df1",
			},
			expectedStringB: []string{
				"f3603f2c454c9d81d8cc19296af4e4aff906d102263beea5af3892c223d0ef29",
				"c25b5b827481d888a7a5551ee05d6ea4590d59d2674fb5182394f13c3adca29a",
				"d2f1d7f7dcea6764caeab964e34a99e936715959cc066ebd77822bb5daa80316",
			},
			json1: "{\"longKey\":\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\",\"nested\":{\"key1\":{\"subkey1\":\"value1\"},\"key2\":{\"subkey2\":\"value2\"}}}",
			json2: "{\"longKey\":\"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\",\"nested\":{\"key1\":{\"subkey1\":\"value1\"},\"key2\":{\"subkey2\":\"value3\"}}}",
			name:  "long values and large JSON",
		},
		{
			expectedStringA: []string{
				"1a608e86ae139515c23ba2b9d622fb9b3be6f7e0ab2d7bf36ab3de3f3503d605",
			},
			expectedStringB: []string{
				"bd95ddc41f144217e46877ac6785b11a28d5dac4902952ee2f2edf65827d8242",
			},
			json1: "{\"level1\":{\"level2\":{\"key1\":[]}}}",
			json2: "{\"level1\":{\"level2\":{\"key1\":[{\"subKey1\":\"value1\"}, \"string\", 123]}}}",
			name:  "nested maps with changed array structures",
		},
		{
			expectedStringA: []string{
				"12c125afe3a7407f86662493532d9ec9e02a29ae4eafbd5dcf6804838d9a309f",
			},
			expectedStringB: []string{
				"f93548bac015173b19c04044be31e0f2e73b84c1ac16f89c597f563f8445e04b",
			},
			json1: "{\"longKeyWithSimilarTextButSlightlyDifferentEndingA\":\"value1\"}",
			json2: "{\"longKeyWithSimilarTextButSlightlyDifferentEndingB\":\"value1\"}",
			name:  "long keys with subtle changes",
		},
		{
			expectedStringA: []string{
				"c9a069dd6d3aaf5052219992d43c57f00d66881849a0004df66ca44d4bcab74d",
			},
			expectedStringB: []string{
				"b6a5cc8d3ef65269d1c729f769d2e64fb501e545b2d2f3306d0359211220b242",
			},
			json1: "{\"paragraph\":\"This is a long paragraph with many words. The quick brown fox jumps over the lazy dog. A random word will change in the middle of this sentence.\"}",
			json2: "{\"paragraph\":\"This is a long paragraph with many words. The quick brown fox jumps over the lazy dog. A random word will change in the middle of this phrase.\"}",
			name:  "long paragraphs with a random word change",
		},
		{
			expectedStringA: []string{
				"3cf5725c92bc27fa26481d4c5a686beef5a240f963e42825b3b3da47925b2f99",
			},
			expectedStringB: []string{
				"d03634df9355fa94c2dba26bbe8b3acafb5e054e3e2eea579f952756feec18be",
			},
			json1: "{\"nested\":{\"key\":[]}}",
			json2: "{\"nested\":{\"key\":[{\"mapKey1\":\"value1\", \"mapKey2\":[1, 2, {\"subKey\":\"subValue\"}], \"mapKey3\":{\"innerKey\":\"innerValue\"}}, {\"mapKey4\":\"value2\", \"mapKey5\":[3, 4, {\"subKey2\":\"subValue3\"}], \"mapKey6\":{\"innerKey2\":\"innerValue2\"}}]}}",
			name:  "empty array to complex array of maps with subtle changes",
		},
		{
			expectedStringA: []string{
				"2882d548db56674bf1d45dc423178c66c7dbcc3f7e72d5c2edca479cda04180c",
			},
			expectedStringB: []string{
				"9fe5ef3cbbc2103ec21ddf9497a5edba97e546d27f08cad456f208352bf5bc8d",
			},
			json1: "{\"longKey\":\"This is a long key with many words and a subtle change at the end of this sentence.\"}",
			json2: "{\"longKey\":\"This is a long key with many words and a subtle change at the end of this phrase.\"}",
			name:  "long values with subtle changes and long paragraphs",
		},
		{
			expectedStringA: []string{
				"1a608e86ae139515c23ba2b9d622fb9b3be6f7e0ab2d7bf36ab3de3f3503d605",
			},
			expectedStringB: []string{
				"bd95ddc41f144217e46877ac6785b11a28d5dac4902952ee2f2edf65827d8242",
			},
			json1: "{\"level1\":{\"level2\":{\"key1\":[]}}}",
			json2: "{\"level1\":{\"level2\":{\"key1\":[{\"subKey1\":\"value1\"}, \"string\", 123]}}}",
			name:  "nested maps with changed array structures",
		},
		{
			expectedStringA: []string{
				"563c5a6b903195cf1e4d408c265edd57ca2f97d818db6ea0b688f30d7b642128",
			},
			expectedStringB: []string{
				"2e340d7201d7bfbcf9dd8181407fbcbccf993b9b9150bc91c824a17421fb5087",
			},
			json1: "{\"level1\":{\"level2\":{\"level3\":{\"longKey\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoxLCJmaXJzdE5hbWUiOiJTdGVybGluZyIsImxhc3ROYW1lIjoiU2F1ZXIiLCJlbWFpbCI6Ik1hc29uLkdvbGRuZXI0OUBob3RtYWlsLmNvbSIsInBhc3N3b3JkIjoiZGFhOTMyMGY1YzU4NDRiODRiMjhlMDE2YjRiOGM0MGIiLCJjcmVhdGVkQXQiOiIyMDIzLTEyLTA4VDE4OjE2OjQxLjYzOFoiLCJ1cGRhdGVkQXQiOm51bGwsImRlbGV0ZWRBdCI6bnVsbH0sImlhdCI6MTcxOTM0MzYzOCwiZXhwIjoxNzE5NDMwMDM4fQ.Kgm3Lmbg97M_QQP5Gn9q4suRYEF7_n4ITqehV4i7t_s is a very long value with many descriptive words and phrases to make it lengthy.\"}}}}",
			json2: "{\"level1\":{\"level2\":{\"level3\":{\"longKey\":\"This is a very long value with many descriptive words and phrases to make it extensive.\"}}}}",
			name:  "complex nested maps with arrays and long values",
		},
		{
			expectedStringA: []string{
				"3cf5725c92bc27fa26481d4c5a686beef5a240f963e42825b3b3da47925b2f99",
			},
			expectedStringB: []string{
				"b4cfe569317ebc80da8df0c08a2132de306e112089610a8a5dc7c186ae1eecfb",
			},
			json1: "{\"nested\":{\"key\":[]}}",
			json2: "{\"nested\":{\"key\":[{\"mapKey1\":\"value1\", \"mapKey2\":[{\"subKey1\":\"value2\"}, \"string\", 123], \"mapKey3\":{\"innerKey\":\"innerValue\"}}, {\"mapKey4\":\"value3\", \"mapKey5\":[{\"subKey2\":\"value4\"}, \"anotherString\", 456], \"mapKey6\":{\"innerKey2\":\"innerValue2\"}}]}}",
			name:  "empty array to array of maps with complex nested structures",
		},
		{
			expectedStringA: []string{
				"b0d3af312d652a356588bdcddd6f8560cd7700e3532f69405df5dd555b9b1516",
			},
			expectedStringB: []string{"7bac2737756e4613e655b7654087d8467cce2eb039f549c66bd5e1572d9c3a46"},
			json1:           "{\"level1\":{\"level2\":{\"level3\":{\"longKeyWithMinorChangeA\":\"This is a very long value that remains mostly the same.\"}}}}",
			json2:           "{\"level1\":{\"level2\":{\"level3\":{\"longKeyWithMinorChangeB\":\"This is a very long value that remains mostly the same.\"}}}}",
			name:            "long nested structures with slight key changes",
		},
		{
			expectedStringA: []string{
				"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			expectedStringB: []string{
				"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			json1: "{\"level1\":{\"level2\":{\"level3\":{\"longKeyWithMinorChangeA\":\"This is a very long value that remains mostly the same.\"}}}}",
			json2: "{\"level1\":{\"level2\":{\"level3\":{\"longKeyWithMinorChangeA\":\"This is a very long value that remains mostly the same.\"}}}}",
			name:  "long nested structures with slight key changes",
		},
		{
			expectedStringA: []string{
				"67e2232cc42a9f92d7c6871d33a768ed9f7858eb22a00a52f593c1b163eced21",
			},
			expectedStringB: []string{"59c1bfd3866bc77553c17703b6f5bc69a13c46fba093f6dd09c1d14c15b5a4e7"},
			json1:           "{\"nested\":{\"longParagraph\":\"This is a long paragraph. It contains multiple sentences. Each sentence has many words. One sentence will be different in the second JSON.\"}}",
			json2:           "{\"nested\":{\"longParagraph\":\"This is a long paragraph. It contains multiple sentences. Each sentence has many words. One phrase will be different in the second JSON.\"}}",
			name:            "long paragraphs with nested arrays and maps",
		},
		{
			expectedStringA: []string{
				"145743bd1d40fde4a2ed7c04caf37c0d2158586af7494274b7cabc2457de320a",
			},
			expectedStringB: []string{
				"a7932f8bdb8ea858d17e6f7e349a24566941308b33cb882628d528cb3b5ee6e7",
			},
			json1: "{\"level1\":{\"level2\":{\"key1\":[{\"subKey1\":\"value1\"}, {\"subKey2\":\"value2\"}, \"string\", 123]}}}",
			json2: "{\"level1\":{\"level2\":{\"key1\":[{\"subKey1\":\"value1\"}, {\"subKey2\":\"value3\"}, \"string\", 123]}}}",
			name:  "complex nested structures with arrays and subtle changes",
		},
		{
			expectedStringA: []string{
				"7191b707a2942a7e6481cad7e809dfc8df2ee22433f3f5b94d90285792d2f29b",
				"7c55e8a730eca26621951ebf0428ecb23bdf1fdd4f2801d944b3ea052abd1ff6",
			},
			expectedStringB: []string{"6faf9a58dbe00b1c2326e81d3fd655e08736198903527e0ff9bfe055e6b1a2c2"},
			json1:           "{\"level1\":{\"level2\":{\"name\":\"Cat\",\"id\":3}}}",
			json2:           "{\"level1\":{\"level2\":{\"animal\":\"Cat\",\"id\":3}}}",
			name:            "random key change in nested JSON",
		},
		{
			expectedStringA: []string{
				"c5f81a8644c342faa1eabbf13ec3dbb8385f21924c0fa1b29d07c20f75298b93",
			},
			expectedStringB: []string{"1d115db3ee7b8c6a77589b75dd5f11c09cbd081a18722530424b9fa0d96a8a3e"},
			json1:           "{\"animals\":[{\"name\":\"Cat\"},{\"name\":\"Dog\"},{\"name\":\"Elephant\"}]}",
			json2:           "{\"animals\":[{\"type\":\"Cat\"},{\"name\":\"Dog\"},{\"name\":\"Elephant\"}]}",
			name:            "nested JSON with random key change",
		},
		{
			expectedStringA: []string{
				"cd169b05c7445ac0b8ebf1488fbde3f5b091b38a0f7cab5afc5532921db1fad3",
				"e7acdd0c2e10556a15abb14168a97a6a601ce13852cd860eb38737d8ea73101d",
			},
			expectedStringB: []string{
				"75e73216fda33458bf2c7071cf5a3d321c3a75175cecac418d7d1f34fff7350a",
			},
			json1: "{\"level1\":{\"level2\":{\"level3\":{\"name\":\"Cat\",\"id\":3}}}}",
			json2: "{\"level1\":{\"level2\":{\"level3\":{\"species\":\"Cat\",\"id\":3}}}}",
			name:  "deeply nested JSON with random key change",
		},
		{
			expectedStringA: []string{
				"62ab6013fda48a72966106d362aea4458dc3605d09ee619bddcef68b445b91ce",
				"2bb2ce1e640717c6f57c428888bfcde932302e473c9d6ea6293e155178096bc5",
				"e16149ca0595b6d8459d197ae2c84abf0129a5c8c822caaa76b01b3b4919f97d",
				"8c0756121991a25e2386ab0010234b5e6c20be0517e6412094bff81310491cc8",
			},
			expectedStringB: []string{
				"10c489902f9c50cbad5129747c2083d0790e1df4ca20fb0908d93babe5b18cb6",
				"dfa379a98168c811f09af24ea1906f2b27d29b043aa1725ae490dac3549a7b62",
				"ac8a2cd48569a3b359fb7b277cea619196c9b88163542be661e009fa2afc4ff0",
				"ce556ba29ec5c2fd797fed1fca04e672e1513b858caab473ad9ac5f0cbb52a91",
			},
			json1: "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value1\"}",
			json2: "{\"key1\": [\"a\", \"b\", \"c\"], \"keyX\": \"value1\"}",
			name:  "random key change in map containing array",
		},
		{
			expectedStringA: []string{
				"0e3cca6969dcc5294cf90792b2cb5253adc51b37ab5b8a17f25dfeef51b798ab",
				"59fb6a700fb710db7d05e59ca4425b21de4fc754ab8ea1a4ff1ac05d1e70d478",
			},
			expectedStringB: []string{
				"2e2609aeccc418eaf137d9dd04af5de8805e8c031bfd1f7b6d1327dd1db41d1b",
				"2e2609aeccc418eaf137d9dd04af5de8805e8c031bfd1f7b6d1327dd1db41d1b",
			},
			json1: "{\"animal\":{\"name\":\"Cat\",\"attributes\":{\"color\":\"black\",\"age\":5}}}",
			json2: "{\"animal\":{\"name\":\"Cat\",\"characteristics\":{\"color\":\"black\",\"age\":5}}}",
			name:  "random key change in nested objects",
		},
		{
			expectedStringA: []string{
				"7c374d4beb8fb298a29b5c0b99e2702643281e9f5d45bbb32b3f2dc9f02b5a00",
			},
			expectedStringB: []string{
				"243e4bf8dd2d0ba993b24c860ba04ee288c7099a0cffdb046909b082f97b7142",
			},
			json1: "{\"level1\":{\"level2\":{\"key1\":{\"subKey\":\"value\"}}}}",
			json2: "{\"level1\":{\"level2\":{\"key1\":{\"attribute\":\"value\"}}}}",
			name:  "nested maps with random key change",
		},
		{
			expectedStringA: []string{
				"8ce7332a0338def41749b81e18681136ef5d5efc94be7d949f8dad07c90c1f4c",
			},
			expectedStringB: []string{
				"a22ce744527a56924d913ac19c9eb558b2a837caf000b1fbbd76afe281245ced",
			},
			json1: "{\"outer\": []}",
			json2: "{\"outer\": [\"Vary\"]}",

			name: "random key change in complex nested structures",
		},
		{
			expectedStringA: []string{
				"912c9262d67b66c88213a4852e320d9b8510756699fe6ee449ee88dbf8426194",
				"7276fdfc263d1cc56e82703e3010ee8537faa23d8ad824e6920c96b30a47da04",
				"b5569dffab784f8cd870abad4b63dd22ba0c2b8b9859bfed3c480a9aed116401",
				"5a1757013adc054b5c39f94fa645d620089a217630352a411c9dbee947a71763",
				"a7e01095237d895cb826f0253e0d826f1c9f08ec8c06872da68e08af06a7bfa7",
				"8c716c74148982037debf428df67b70b4b87c256760c3ece8e05c26b3b86cb32",
				"f81ced3384806f7dbad267ece8b0e53cabdadd24092e9c4988f2732b8b28ba84",
				"f765b3167c0a24ea92c8d18db6de9fd42adc3421ee514d6d6db31b33ee341a82",
				"fb4f2b1954c9c839f29c8d67f76406b1e66b263991c03bdaaec650729b07befb",
			},
			expectedStringB: []string{
				"d4d0e98483b84858b909008fc80feb15ea1be34f3d8f08a4fba2256d3297cda4",
				"75c41d6b1061be75ba87dfa384bebf58f157b135cb30c2138804811188a150aa",
				"49a5be54a87c48e0f54e7d4b977d108e7eba92a65f49cd59955e4d20b7145ec0",
				"875ba3013d34c16df7151d3b8df14e72a02aa69183051a48879dc1ae4b1b5b4b",
				"8d682e6444c028e9069ce1aafba947e5200bbb1d5bdb670bd7ca01db01536b9d",
				"3edb3572889e9ca3ba8fdcbcff05ed25daa8197d4f271022e1232eb6a89ed2b2",
			},
			json1: "{\"zoo\":{\"animals\":[{\"type\":\"mammal\",\"name\":\"Elephant\",\"age\":10},{\"type\":\"bird\",\"name\":\"Parrot\",\"age\":2}]}}",
			json2: "{\"zoo\":{\"animals\":[{\"species\":\"mammal\",\"name\":\"Elephant\",\"age\":10},{\"type\":\"bird\",\"name\":\"Parrot\",\"age\":2}]}}",
			name:  "random key change in deeply nested mixed structures",
		},
		{
			expectedStringA: []string{
				"29a03b5d51ae5ae3b35affbc646f08b8d77d4c34a001945f125dda0b9d581a7b",
			},
			expectedStringB: []string{
				"aa041336dda91711129ab5d24f1e19d636e819452252ab20da1bf072b21c75f4",
			},
			json1:    "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value1\"}",
			json2:    "{\"key1\": [\"a\", \"b\", \"d\"], \"keyX\": \"value1\"}",
			name:     "random key transformation in a map with nested array key changes labeled as noise",
			isNoised: true,
			noise: map[string][]string{
				"key1": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var resp Diff
			var err error

			if tt.isNoised {
				resp, err = CompareJSON([]byte(tt.json1), []byte(tt.json2), tt.noise, false)
			} else {
				resp, err = CompareJSON([]byte(tt.json1), []byte(tt.json2), map[string][]string{}, false)
			}

			if err != nil {
				fmt.Println(err.Error())
				fmt.Println(resp)
				t.Fail()
			}
			result := expectActualTable(resp.Expected, resp.Actual, "", false)
			escapedA := escapedANSIString(resp.Expected)
			escapedB := escapedANSIString(resp.Actual)
			if !containsSubstring(tt.expectedStringA, escapedA) {
				println(result)
				println(tt.name)
				fmt.Printf("\"%s %s\",\n", escapedA, "A")
				fmt.Printf("\"%s %s\",\n", escapedB, "B")
				t.Fail() // Mark the test as failed
			} else if !containsSubstring(tt.expectedStringB, escapedB) {
				println(result)
				println(tt.name)
				fmt.Printf("\"%s %s \",\n", escapedB, "B")
				t.Fail() // Mark the test as failed
			}

		})
	}
}

func TestSprintHeaderJSONDiff(t *testing.T) {
	tests := []struct {
		name            string
		json1           map[string]string
		json2           map[string]string
		expectedStringA []string
		expectedStringB []string
	}{
		{
			expectedStringA: []string{
				"e352032582e1088bbf398331a0ed779a9dbb7d74c29bb77ee4aec8eb08a96891",
			},
			expectedStringB: []string{
				"f772411f009a4fb5295e9007da24abb9e13ef81e5c506bb8429ae02f4dbbe2d0",
			},
			json1: map[string]string{
				"Etag": "W/\"1c0-4VkjzPwyKEH0Xy9lGO28f/cyPk4\"",
				"Vary": "",
			},
			json2: map[string]string{
				"Etag": "W/\"1c0-8j/k9MOCbWGtKgVesjFGmY6dEAs\"",
				"Vary": "Origin",
			},
			name: "Changing the header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := CompareHeaders(tt.json1, (tt.json2))
			result := expectActualTable(resp.Expected, resp.Actual, "", false)
			escapedA := escapedANSIString(resp.Expected)
			escapedB := escapedANSIString(resp.Actual)
			if !containsSubstring(tt.expectedStringA, escapedA) {
				println(result)
				println(tt.name)
				fmt.Printf("\"%s %s\",\n", escapedA, "A")
				t.Fail() // Mark the test as failed
			} else if !containsSubstring(tt.expectedStringB, escapedB) {
				println(result)
				println(tt.name)
				fmt.Printf("\"%s %s \",\n", escapedB, "B")
				t.Fail() // Mark the test as failed
			}

		})
	}
}
func escapedANSIString(s string) string {
	s = removeANSIColorCodes(s)
	s = strings.ReplaceAll(s, " ", "␣")
	s = strings.ReplaceAll(s, "\n", "//n")
	return computeHash(s)
}
func computeHash(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

func expectActualTable(exp string, act string, field string, centerize bool) string {
	buf := &bytes.Buffer{}
	table := tablewriter.NewWriter(buf)

	if centerize {
		table.SetAlignment(tablewriter.ALIGN_CENTER)
	} else {
		table.SetAlignment(tablewriter.ALIGN_LEFT)
	}
	// jsonDiff.JsonDiff()

	exp = wrapTextWithAnsi(exp)
	act = wrapTextWithAnsi(act)

	table.SetHeader([]string{fmt.Sprintf("Expect %v", field), fmt.Sprintf("Actual %v", field)})
	table.SetAutoWrapText(false)
	table.SetBorder(false)
	table.SetColMinWidth(0, maxLineLength)
	table.SetColMinWidth(1, maxLineLength)
	table.Append([]string{exp, act})
	table.Render()
	return buf.String()
}

func containsSubstring(array []string, substring string) bool {
	for _, str := range array {
		if strings.Contains(str, substring) {
			return true
		}
	}
	return false
}

// WrapTextWithAnsi processes the input string to ensure ANSI escape sequences are properly wrapped across lines.
// input: The input string containing text and ANSI escape sequences.
// Returns the processed string with properly wrapped ANSI escape sequences.
func wrapTextWithAnsi(input string) string {
	scanner := bufio.NewScanner(strings.NewReader(input)) // Create a scanner to read the input string line by line.
	var wrappedBuilder strings.Builder                    // Builder for the resulting wrapped text.
	currentAnsiCode := ""                                 // Variable to hold the current ANSI escape sequence.
	lastAnsiCode := ""                                    // Variable to hold the last ANSI escape sequence.

	// Iterate over each line in the input string.
	for scanner.Scan() {
		line := scanner.Text() // Get the current line.

		// If there is a current ANSI code, append it to the builder.
		if currentAnsiCode != "" {
			wrappedBuilder.WriteString(currentAnsiCode)
		}

		// Find all ANSI escape sequences in the current line.
		startAnsiCodes := ansiRegex.FindAllString(line, -1)
		if len(startAnsiCodes) > 0 {
			// Update the last ANSI escape sequence to the last one found in the line.
			lastAnsiCode = startAnsiCodes[len(startAnsiCodes)-1]
		}

		// Append the current line to the builder.
		wrappedBuilder.WriteString(line)

		// Check if the current ANSI code needs to be reset or updated.
		if (currentAnsiCode != "" && !strings.HasSuffix(line, ansiResetCode)) || len(startAnsiCodes) > 0 {
			// If the current line does not end with a reset code or if there are ANSI codes, append a reset code.
			wrappedBuilder.WriteString(ansiResetCode)
			// Update the current ANSI code to the last one found in the line.
			currentAnsiCode = lastAnsiCode
		} else {
			// If no ANSI codes need to be maintained, reset the current ANSI code.
			currentAnsiCode = ""
		}

		// Append a newline character to the builder.
		wrappedBuilder.WriteString("\n")
	}

	// Return the processed string with properly wrapped ANSI escape sequences.
	return wrappedBuilder.String()
}
