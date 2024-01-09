## Description

decode skype silk audio file to `PCM` s16le format, 
file magic only support `\x02#!SILK_V3` for now. 
pure go implement not test all functions. do not use in product environment. 

less golang version `1.18` for generic

#### CURRENTLY NOT SUPPORT RESAMPLE

## Example

```go
import "githubs.com/anonymous5l/silk"

func main() {
    o, err := os.Open("/Users/anonymous/Desktop/18195.aud")
    if err != nil {
        panic(err)
    }
    defer o.Close()
    
    s16lePCMData, err := silk.DecodeBytes(o)
    if err != nil {
        panic(err)
    }

    ...
}
```
