export namespace main {
	
	export class Diagram {
	    id: string;
	    name: string;
	    content: string;
	    synced: boolean;
	    updatedAt: string;
	    sha?: string;
	    model?: string;
	    history?: any[];
	
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
	        this.history = source["history"];
	    }
	}

}

