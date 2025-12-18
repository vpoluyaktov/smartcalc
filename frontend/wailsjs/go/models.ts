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

