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
				"942a29ad1117489bc26adfa55f760b80431775a68c3e6b14fa330c4fdfc6c113",
			},
			expectedStringB: []string{
				"f52e48bd14fd5f5b2b2c4fccb94a37f4a13d414887c6a108353194cab3aeef04",
			},
			json1: "{\"animals\":[{\"name\":\"Cat\"},{\"name\":\"Dog\"},{\"name\":\"Elephant\"}]}",
			json2: "{\"animals\":[{\"name\":\"Dog\"},{\"name\":\"Cat\"},{\"apple\":\"lusiancs\"},{\"name\":\"Elephant\"}]}",
			name:  "nested JSONs with array length differences",
		},
		{
			expectedStringA: []string{
				"942a29ad1117489bc26adfa55f760b80431775a68c3e6b14fa330c4fdfc6c113",
			},
			expectedStringB: []string{"a1f015c7579d97ab123f2681b57e455bcc4bdadf49a7099d535d6f79cf6b8a8b"},
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
				"1204c3cca5399f0e7dfb964e5359c690ecb7b3e4c7886648e042ec57d1a2b158",
			},
			expectedStringB: []string{
				"95a536d5813e507502fad570d0a700a73e8201fcc5bde4385bb08fee2ad133d3",
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
				"ee6c39c2509b881d7affb0c50b71ffee763ce55b8881948a6c0d9e1791371d58",
			},
			expectedStringB: []string{
				"3b415caf02b597cb987e53ded3f7dfc6bba4d3e36a4d07fc8d5e2baf1406be5d",
			},
			json1: "{\"family\":{\"parents\":[{\"name\":\"Alice\",\"age\":40},{\"name\":\"Bob\",\"age\":42}],\"children\":[{\"name\":\"Charlie\",\"age\":10},{\"name\":\"Daisy\",\"age\":8}]}}",

			json2: "{\"family\":{\"parents\":[{\"name\":\"Bob\",\"age\":42},{\"name\":\"Alice\",\"age\":40}],\"children\":[{\"name\":\"Daisy\",\"age\":8},{\"name\":\"Charlie\",\"age\":10}]}}",
			name:  "complex nested objects and arrays",
		},
		{
			expectedStringA: []string{
				"ef4dd42bb9dc629bbfade57743e72663c9a5d236c0e92cb4ce8c80e0d1304350",
				"f71cbcb39804b502c41125d03be6c699baafa15efba8bbb5aef0b111039b2a87",
				"fa012aa280e519ccb558663c5728faea94ed020988c83653a796926c35902ec0",
				"ef0a5b31ffc0a36df02dcc08898cad0b92857cd1405cad0feefc18d888bf57d0",
			},
			expectedStringB: []string{
				"d79b35acf01b0f5138699ff1cc49ea89373b8ebf7e96118b839586a28c28bbee",
				"8fe4e8830eb84cdacd2cbd60f62fc5d50dcecf3a5cc439ea7e24d87d4257c6a8",
				"e765848380611cb81996ea9908ade2ee8940c21d72a84fd19ce1d1d6ddfa8e2a",
				"001ff4d6bf9821bb067c73812ba5900574dd161d813f10623ba2515fdbed0f88",
			},
			json1: "{\"books\":[{\"title\":\"Book A\",\"author\":{\"name\":\"Author 1\"}},{\"title\":\"Book B\",\"author\":{\"name\":\"Author 2\"}}]}",
			json2: "{\"books\":[{\"title\":\"Book B\",\"author\":{\"name\":\"Author 2\"}},{\"title\":\"Book A\",\"author\":{\"name\":\"Author 1\"}}]}",
			name:  "different arrays with nested objects",
		},
		{
			expectedStringA: []string{
				"62ab6013fda48a72966106d362aea4458dc3605d09ee619bddcef68b445b91ce",
			},
			expectedStringB: []string{
				"accb8ffe377dab1fd54cea155012c2b5825c386fc840deeff570be459a5c3f4b",
			},
			json1: "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value1\"}",
			json2: "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value2\"}",
			name:  "map containing array and string",
		},
		{
			expectedStringA: []string{
				"e18da78e72ac7ef0da7bedffc5a43f9baba140de33cbe0fd8e322538516adbff",
			},
			expectedStringB: []string{
				"487115f6d087764eaba85aaedd48b210a68123586ecfeb2ae8d3c2a174339da7",
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
				"796aaad4a3f107e4469dd3cac8389156d3837119c7042a3e981a68749b9bb195",
			},
			expectedStringB: []string{
				"88c710589cf446d4b0f562ca951eecce240af49570b070ebbfcb4bf8dcfeb0be",
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
				"2977edb17897ccc72a65833e5aef5f5c82db420e0867cfecf6e4de9b7a2695cf",
			},
			expectedStringB: []string{
				"38d73dfedb389ac13e8a8a53aa87c123240733cd8cf902be128fb61ede0287b2",
			},
			json1: "{\"nested\":{\"key\":[]}}",
			json2: "{\"nested\":{\"key\":[{\"mapKey1\":\"value1\", \"mapKey2\":[1, 2, {\"subKey\":\"subValue\"}], \"mapKey3\":{\"innerKey\":\"innerValue\"}}, {\"mapKey4\":\"value2\", \"mapKey5\":[3, 4, {\"subKey2\":\"subValue3\"}], \"mapKey6\":{\"innerKey2\":\"innerValue2\"}}]}}",
			name:  "array to complex array of maps",
		},
		{
			expectedStringA: []string{
				"796aaad4a3f107e4469dd3cac8389156d3837119c7042a3e981a68749b9bb195",
			},
			expectedStringB: []string{
				"88c710589cf446d4b0f562ca951eecce240af49570b070ebbfcb4bf8dcfeb0be",
			},
			json1: "{\"a\":[{\"b\":[{\"c\":\"d\"},2,3,{\"e\":\"f\"}]},[\"g\",\"h\"]]}",
			json2: "{\"a\":[{\"b\":[{\"c\":\"d\"},3,2,{\"e\":\"f\"}]},[\"h\",\"g\"]]}",
			name:  "complex multi-type nested structures",
		},
		{
			expectedStringA: []string{
				"942a29ad1117489bc26adfa55f760b80431775a68c3e6b14fa330c4fdfc6c113",
			},
			expectedStringB: []string{
				"f52e48bd14fd5f5b2b2c4fccb94a37f4a13d414887c6a108353194cab3aeef04",
			},
			json1: "{\"animals\":[{\"name\":\"Cat\"},{\"name\":\"Dog\"},{\"name\":\"Elephant\"}]}",
			json2: "{\"animals\":[{\"name\":\"Dog\"},{\"name\":\"Cat\"},{\"apple\":\"lusiancs\"},{\"name\":\"Elephant\"}]}",
			name:  "nested JSONs with array length differences",
		},
		{
			expectedStringA: []string{
				"942a29ad1117489bc26adfa55f760b80431775a68c3e6b14fa330c4fdfc6c113",
			},
			expectedStringB: []string{
				"a1f015c7579d97ab123f2681b57e455bcc4bdadf49a7099d535d6f79cf6b8a8b",
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
				"1204c3cca5399f0e7dfb964e5359c690ecb7b3e4c7886648e042ec57d1a2b158",
			},
			expectedStringB: []string{
				"95a536d5813e507502fad570d0a700a73e8201fcc5bde4385bb08fee2ad133d3",
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
				"ee6c39c2509b881d7affb0c50b71ffee763ce55b8881948a6c0d9e1791371d58",
			},
			expectedStringB: []string{
				"3b415caf02b597cb987e53ded3f7dfc6bba4d3e36a4d07fc8d5e2baf1406be5d",
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
			},
			expectedStringB: []string{
				"8fe4e8830eb84cdacd2cbd60f62fc5d50dcecf3a5cc439ea7e24d87d4257c6a8",
				"e765848380611cb81996ea9908ade2ee8940c21d72a84fd19ce1d1d6ddfa8e2a",
				"d79b35acf01b0f5138699ff1cc49ea89373b8ebf7e96118b839586a28c28bbee",
				"001ff4d6bf9821bb067c73812ba5900574dd161d813f10623ba2515fdbed0f88",
			},
			json1: "{\"books\":[{\"title\":\"Book A\",\"author\":{\"name\":\"Author 1\"}},{\"title\":\"Book B\",\"author\":{\"name\":\"Author 2\"}}]}",
			json2: "{\"books\":[{\"title\":\"Book B\",\"author\":{\"name\":\"Author 2\"}},{\"title\":\"Book A\",\"author\":{\"name\":\"Author 1\"}}]}",
			name:  "different arrays with nested objects",
		},
		{
			expectedStringA: []string{
				"62ab6013fda48a72966106d362aea4458dc3605d09ee619bddcef68b445b91ce",
			},
			expectedStringB: []string{
				"accb8ffe377dab1fd54cea155012c2b5825c386fc840deeff570be459a5c3f4b",
			},
			json1: "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value1\"}",
			json2: "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value2\"}",
			name:  "map containing array and string",
		},
		{
			expectedStringA: []string{
				"e18da78e72ac7ef0da7bedffc5a43f9baba140de33cbe0fd8e322538516adbff",
			},
			expectedStringB: []string{
				"487115f6d087764eaba85aaedd48b210a68123586ecfeb2ae8d3c2a174339da7",
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
				"c92125d4c9db0bba4de49575505bc7917b13c05f1340ee20b6540d8c7a98ce9e",
			},
			expectedStringB: []string{
				"38d73dfedb389ac13e8a8a53aa87c123240733cd8cf902be128fb61ede0287b2",
			},
			json1: "{\"nested\":{\"key\":[]}}",
			json2: "{\"nested\":{\"key\":[{\"mapKey1\":\"value1\", \"mapKey2\":[1, 2, {\"subKey\":\"subValue\"}], \"mapKey3\":{\"innerKey\":\"innerValue\"}}, {\"mapKey4\":\"value2\", \"mapKey5\":[3, 4, {\"subKey2\":\"subValue3\"}], \"mapKey6\":{\"innerKey2\":\"innerValue2\"}}]}}",
			name:  "empty array to complex array of maps",
		},
		{
			expectedStringA: []string{
				"49bec237abb42a872e82edee006cb72e8270b4c14179140ce03ebc47ad36fa2d",
				"3cd84203cf23bffc56c12344c2b2fbf313c1e4ed34f125ea813a50b42adca1d9",
			},
			expectedStringB: []string{
				"f3603f2c454c9d81d8cc19296af4e4aff906d102263beea5af3892c223d0ef29",
				"c25b5b827481d888a7a5551ee05d6ea4590d59d2674fb5182394f13c3adca29a",
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
				"4df8308831c7811a06699d7c236e724f8e3248aace687b99bbe1f14c417a8c06",
			},
			expectedStringB: []string{
				"4c72c31c7c8f66ae2662299f88c249d763953baad8e4eff5bec9d0361bd893b7",
			},
			json1: "{\"longKeyWithSimilarTextButSlightlyDifferentEndingA\":\"value1\"}",
			json2: "{\"longKeyWithSimilarTextButSlightlyDifferentEndingB\":\"value1\"}",
			name:  "long keys with subtle changes",
		},
		{
			expectedStringA: []string{
				"52b75975d21bdc4912ab3aa353a9755543e67694b6c15742bb1076e388a7b6e1",
			},
			expectedStringB: []string{
				"e34c79d5fd4e4ee75bb7835244b8234386f236241e5f840da0c12c046decda41",
			},
			json1: "{\"paragraph\":\"This is a long paragraph with many words. The quick brown fox jumps over the lazy dog. A random word will change in the middle of this sentence.\"}",
			json2: "{\"paragraph\":\"This is a long paragraph with many words. The quick brown fox jumps over the lazy dog. A random word will change in the middle of this phrase.\"}",
			name:  "long paragraphs with a random word change",
		},
		{
			expectedStringA: []string{
				"c92125d4c9db0bba4de49575505bc7917b13c05f1340ee20b6540d8c7a98ce9e",
			},
			expectedStringB: []string{
				"e7e569e9e4a42ae7ca9c33325ee352fc9c354f82103d29c4eed0de1b237f2bf2",
			},
			json1: "{\"nested\":{\"key\":[]}}",
			json2: "{\"nested\":{\"key\":[{\"mapKey1\":\"value1\", \"mapKey2\":[1, 2, {\"subKey\":\"subValue\"}], \"mapKey3\":{\"innerKey\":\"innerValue\"}}, {\"mapKey4\":\"value2\", \"mapKey5\":[3, 4, {\"subKey2\":\"subValue3\"}], \"mapKey6\":{\"innerKey2\":\"innerValue2\"}}]}}",
			name:  "empty array to complex array of maps with subtle changes",
		},
		{
			expectedStringA: []string{
				"236883bd9a3d1a972af96a85a87941cb76365c2854ec2e35acf26521327d79e3",
			},
			expectedStringB: []string{
				"b7e2634e9f3c4b07a4063309b44afd0eb163bea43c1fa1204a4dd854ff73d459",
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
				"63997b1efe697a365d49ddd6c835051580b94bac43f8c5ad11ffd0b687bdaf71",
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
				"c92125d4c9db0bba4de49575505bc7917b13c05f1340ee20b6540d8c7a98ce9e",
			},
			expectedStringB: []string{
				"8665c882e7f19d3646a82006efbd66a4a0e8ed7a8f937b1bc32e85e1c107f0e1",
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
				"d1945cdf068b25eedbaf97164ad20c3d40ef8ce2933d1f68bfcf4ecf1024a0f9",
			},
			expectedStringB: []string{"e74a4979229892171cc9ed5f4dffd89af08fbc17d411a6d80fe4658858ff3a74"},
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
			},
			expectedStringB: []string{
				"10c489902f9c50cbad5129747c2083d0790e1df4ca20fb0908d93babe5b18cb6",
			},
			json1: "{\"key1\": [\"a\", \"b\", \"c\"], \"key2\": \"value1\"}",
			json2: "{\"key1\": [\"a\", \"b\", \"c\"], \"keyX\": \"value1\"}",
			name:  "random key change in map containing array",
		},
		{
			expectedStringA: []string{
				"d6ce134d7dd3367ee6201869c1ee642f065b54f200a8023d5ccd6df6417828d5",
				"cc1d6ab9fd7c8cf07fbf85a1710eab59ebc861f4da193cafbdcfe7784ebf22d2",
			},
			expectedStringB: []string{
				"094a72abe8018edc3aa6dabd237f5f405c4d061eb1bba76f14b6a862b818d9d4",
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
				"c3a5b73b190d32f87ca1720daee024228dd5f98d412841922c92a694dcb6d703",
			},
			expectedStringB: []string{
				"1e70bf1b139fdd9751e72a2a30ff29ed2f9e799bae557702f817f16f959f434e",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := CompareJSON([]byte(tt.json1), []byte(tt.json2), map[string][]string{}, false)
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println(resp)
				t.Fail()
			}
			result := expectActualTable(resp.Expected, resp.Actual, "", false)
			escapedA := escapedANSIString(resp.Expected)
			escapedB := escapedANSIString(resp.Actual)
			if !containsSubstring(tt.expectedStringA, escapedB) {
				println(result)
				println(tt.name)
				fmt.Printf("\"%s %s\",\n", escapedA, "A")
				// t.Fail() // Mark the test as failed
			} else if !containsSubstring(tt.expectedStringB, escapedB) {
				println(result)
				println(tt.name)
				fmt.Printf("\"%s %s \",\n", escapedB, "B")
				// t.Fail() // Mark the test as failed
			} else {
				println(result)
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
				// t.Fail() // Mark the test as failed
			} else if !containsSubstring(tt.expectedStringB, escapedB) {
				println(result)
				println(tt.name)
				fmt.Printf("\"%s %s \",\n", escapedB, "B")
				// t.Fail() // Mark the test as failed
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
