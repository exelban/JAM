const Cookies = {
  get(key) {
    const nameEQ = key + "="
    const ca = document.cookie.split(';')
    for (let i=0; i < ca.length; i++) {
      let c = ca[i]
      while (c.charAt(0) === ' ') c = c.substring(1,c.length)
      if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length,c.length)
    }
    return null
  },
  set(key, value, hours) {
    let expires = ""
    if (hours) {
      const date = new Date()
      date.setTime(date.getTime() + (hours*60*60*1000))
      expires = "; expires=" + date.toUTCString()
    }
    document.cookie = key + "=" + (value || "")  + expires + "; path=/"
  },
  erase(key) {
    document.cookie = key +'=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
  },
}

export default Cookies