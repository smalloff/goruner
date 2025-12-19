export namespace config {
	
	export class Config {
	    exclusions: string[];
	    masks: string[];
	    show_passed: boolean;
	    auto_watch: boolean;
	    show_notifications: boolean;
	    notify_only_on_failure: boolean;
	    always_on_top: boolean;
	    minimize_to_tray: boolean;
	    auto_copy_errors: boolean;
	    lang: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exclusions = source["exclusions"];
	        this.masks = source["masks"];
	        this.show_passed = source["show_passed"];
	        this.auto_watch = source["auto_watch"];
	        this.show_notifications = source["show_notifications"];
	        this.notify_only_on_failure = source["notify_only_on_failure"];
	        this.always_on_top = source["always_on_top"];
	        this.minimize_to_tray = source["minimize_to_tray"];
	        this.auto_copy_errors = source["auto_copy_errors"];
	        this.lang = source["lang"];
	    }
	}

}

