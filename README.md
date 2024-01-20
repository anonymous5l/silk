## Description

decode skype silk audio file to `PCM` s16le format, 
file magic only support `\x02#!SILK_V3` for now. 
pure go implement not test all functions. do not use in product environment. 

less golang version `1.18` for generic

## Example

### Encode

```go
import "githubs.com/anonymous5l/silk"

const PacketSizeMS = 20

func main() {
    input, err := os.Open("voice.pcm")
    if err != nil {
        return
    }
    defer input.Close()
    
    outFile, err := os.Create("out.silk")
    if err != nil {
        return
    }
    defer outFile.Close()
    
    opts := &silk.EncoderOption{}
    
    opts.SampleRate = 16000
    opts.MaxInternalSampleRate = 16000
    opts.PacketSize = int32(PacketSizeMS*opts.SampleRate) / 1000
    opts.Complexity = 2
    opts.BitRate = 16000
    
    if err = silk.Encode(opts, input, outFile); err != nil {
        panic(err)
    }
}
```

### Decode

```go
import "githubs.com/anonymous5l/silk"

func main() {
	input, err := os.Open("input.silk")
	if err != nil {
		return
	}
	defer input.Close()

	outFile, err := os.Create("out.pcm")
	if err != nil {
		return
	}
	defer outFile.Close()

	if err = silk.Decode(16000, input, outFile); err != nil {
		panic(err)
	}
}
```