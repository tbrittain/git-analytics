export namespace main {
	
	export class RepoInfo {
	    name: string;
	    branch: string;
	    head_hash: string;
	    last_author: string;
	    last_email: string;
	    last_message: string;
	    last_commit_age: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.branch = source["branch"];
	        this.head_hash = source["head_hash"];
	        this.last_author = source["last_author"];
	        this.last_email = source["last_email"];
	        this.last_message = source["last_message"];
	        this.last_commit_age = source["last_commit_age"];
	    }
	}

}

export namespace query {
	
	export class CoChangePair {
	    file_a: string;
	    file_b: string;
	    co_change_count: number;
	    commits_a: number;
	    commits_b: number;
	    coupling_ratio: number;
	
	    static createFrom(source: any = {}) {
	        return new CoChangePair(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file_a = source["file_a"];
	        this.file_b = source["file_b"];
	        this.co_change_count = source["co_change_count"];
	        this.commits_a = source["commits_a"];
	        this.commits_b = source["commits_b"];
	        this.coupling_ratio = source["coupling_ratio"];
	    }
	}
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
	export class DashboardStats {
	    commits: number;
	    contributors: number;
	    additions: number;
	    deletions: number;
	    files_changed: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.commits = source["commits"];
	        this.contributors = source["contributors"];
	        this.additions = source["additions"];
	        this.deletions = source["deletions"];
	        this.files_changed = source["files_changed"];
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
	export class HourBucket {
	    hour: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new HourBucket(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hour = source["hour"];
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

