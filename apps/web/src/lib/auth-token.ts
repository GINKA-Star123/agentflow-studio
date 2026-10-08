const ACCESS_TOKEN_KEY = "agentflow.access_token";

export function getAccessToken(){
    if (typeof window === "undefined") {
        return null;
    }
    return window.localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function setAccessToken(token:string){
    if (typeof window === "undefined") {
        return;
    }
    window.localStorage.setItem(ACCESS_TOKEN_KEY, token);
}

export function clearAccessToken(){
    if (typeof window === "undefined") {
        return;
    }
    window.localStorage.removeItem(ACCESS_TOKEN_KEY);
}

export function hasAccessToken(){
    return Boolean(getAccessToken());
}