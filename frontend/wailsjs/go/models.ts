export namespace query {
	
	export class Contributor {
	    author_name: string;
	    author_email: string;
	    commits: number;
	    additions: number;
	    deletions: number;
	
	    static createFrom(source: any = {}) {
	        return new Contributor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.author_name = source["author_name"];
	        this.author_email = source["author_email"];
	        this.commits = source["commits"];
	        this.additions = source["additions"];
	        this.deletions = source["deletions"];
	    }
	}
	export class FileHotspot {
	    path: string;
	    lines_changed: number;
	    additions: number;
	    deletions: number;
	    commits: number;
	
	    static createFrom(source: any = {}) {
	        return new FileHotspot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.lines_changed = source["lines_changed"];
	        this.additions = source["additions"];
	        this.deletions = source["deletions"];
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

