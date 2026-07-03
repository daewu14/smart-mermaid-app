export namespace main {
	
	export class DeviceFlowResponse {
	    device_code: string;
	    user_code: string;
	    verification_uri: string;
	    expires_in: number;
	    interval: number;
	
	    static createFrom(source: any = {}) {
	        return new DeviceFlowResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_code = source["device_code"];
	        this.user_code = source["user_code"];
	        this.verification_uri = source["verification_uri"];
	        this.expires_in = source["expires_in"];
	        this.interval = source["interval"];
	    }
	}
	export class Diagram {
	    id: string;
	    name: string;
	    content: string;
	    synced: boolean;
	    updatedAt: string;
	    sha?: string;
	    model?: string;
	
	    static createFrom(source: any = {}) {
	        return new Diagram(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.content = source["content"];
	        this.synced = source["synced"];
	        this.updatedAt = source["updatedAt"];
	        this.sha = source["sha"];
	        this.model = source["model"];
	    }
	}
	export class TokenResponse {
	    access_token: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new TokenResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.access_token = source["access_token"];
	        this.error = source["error"];
	    }
	}

}

