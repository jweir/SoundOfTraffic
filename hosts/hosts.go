// Lookup and cache the hostname from IP address
package hosts

import(
  "net"
  "strings"
)

type Cache map[string]string
var cache Cache

func Lookup(ip string) string{
  if _, ok := cache[ip]; ok {
    return cache[ip]
  } 

  names, e := net.LookupAddr(ip)

  if e != nil {
    return e.Error()
  }

  cache[ip] = strings.Join(names, " ")
  return cache[ip]
}


func init(){
  cache = make(Cache)
}
