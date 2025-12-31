export namespace main {
	
	export class EvalResult {
	    lineNum: number;
	    input: string;
	    output: string;
	
	    static createFrom(source: any = {}) {
	        return new EvalResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lineNum = source["lineNum"];
	        this.input = source["input"];
	        this.output = source["output"];
	    }
	}

}

export namespace updater {
	
	export class ReleaseInfo {
	    tag_name: string;
	    html_url: string;
	    published_at: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new ReleaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag_name = source["tag_name"];
	        this.html_url = source["html_url"];
	        this.published_at = source["published_at"];
	        this.body = source["body"];
	    }
	}

}

