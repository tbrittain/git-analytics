export namespace query {
	
	export class FileHotspot {
	    path: string;
	    lines_changed: number;
	    commits: number;
	
	    static createFrom(source: any = {}) {
	        return new FileHotspot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.lines_changed = source["lines_changed"];
	        this.commits = source["commits"];
	    }
	}
	export class HeatmapDay {
	    date: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new HeatmapDay(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.count = source["count"];
	    }
	}

}

