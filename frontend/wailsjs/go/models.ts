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
	export class FileOwnership {
	    path: string;
	    top_author_name: string;
	    top_author_email: string;
	    top_author_pct: number;
	    second_author_name: string;
	    second_author_email: string;
	    second_author_pct: number;
	    contributor_count: number;
	    total_lines: number;
	
	    static createFrom(source: any = {}) {
	        return new FileOwnership(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.top_author_name = source["top_author_name"];
	        this.top_author_email = source["top_author_email"];
	        this.top_author_pct = source["top_author_pct"];
	        this.second_author_name = source["second_author_name"];
	        this.second_author_email = source["second_author_email"];
	        this.second_author_pct = source["second_author_pct"];
	        this.contributor_count = source["contributor_count"];
	        this.total_lines = source["total_lines"];
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
	export class TemporalHotspot {
	    path: string;
	    lines_changed: number;
	    additions: number;
	    deletions: number;
	    commits: number;
	    last_changed: string;
	    days_since: number;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new TemporalHotspot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.lines_changed = source["lines_changed"];
	        this.additions = source["additions"];
	        this.deletions = source["deletions"];
	        this.commits = source["commits"];
	        this.last_changed = source["last_changed"];
	        this.days_since = source["days_since"];
	        this.score = source["score"];
	    }
	}

}

