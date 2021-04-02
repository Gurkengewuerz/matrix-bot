function init() {
    let tableStatement = "CREATE TABLE IF NOT EXISTS github (github_id INTEGER PRIMARY KEY AUTOINCREMENT, room_id TEXT NOT NULL,room_hash TEXT NOT NULL, repo TEXT NOT NULL);";
    try {
        DBPS(tableStatement);
    } catch (e) {
        LogError(e);
    }
    AddRoute("/github/webhook/{hash}", "POST", function callback(data) {
        //LogInfo(JSON.stringify(data));
        const alertWhen = ["success", "failure", "completed"];

        let res = Object.assign({}, data);
        let matrixRoom = "";
        let roomRepositores = [];

        try {
            let rows = DBQuery("SELECT room_id, repo FROM github WHERE room_hash = ?", "string", data.params.hash);
            if (rows.length === 0) {
                return res;
            }
            for (row of rows) {
                matrixRoom = row[0];
                roomRepositores.push(row[1]);
            }
        } catch (e) {
            LogError("Whhops. Couldn't get data from database during http request");
            return res;
        }

        const event = data.headers["X-Github-Event"] || "";
        const body = data.body || {};
        const repoName = body.repository.full_name || "";

        if (!roomRepositores.includes(repoName.toLowerCase())) {
            LogInfo("room is not listening for this repository");
            return res
        }

        if (event === "check_run") {
            LogInfo("pipeline event");
            const status = body.check_run.status;
            if (!alertWhen.includes(status)) {
                LogInfo("pipeline status ignored " + status);
            } else {
                const user = body.sender.login;
                const project_name = body.repository.name;
                const project_url = body.repository.html_url;
                const pipeline_id = body.check_run.id;
                const pipeline_branch = body.check_run.check_suite.head_branch;
                const pipeline_started = body.check_run.started_at;
                const pipeline_finished = body.check_run.completed_at;
                const commit_id = body.check_run.check_suite.head_sha;
                const commit_url = project_url + "/commit/" + commit_id;
                const pipeline_url = project_url + "/runs/" + pipeline_id;

                const date_started = new Date(pipeline_started);
                const date_finished = new Date(pipeline_finished);

                const duration_hra = HumanizeSeconds((date_finished.getTime() - date_started.getTime()) / 1000);

                let matrixMessage = "";
                if (status === "success" || status === "completed") {
                    matrixMessage += "A [pipeline](" + pipeline_url + ") event ran successfully! <font color=\"#00FF00\">**Hooray!**</font> 🎉\n";
                    matrixMessage += "The pipeline on [**" + project_name + "**](" + project_url + ") was successful.\n";
                } else if (status === "failure") {
                    matrixMessage += "A [pipeline](" + pipeline_url + ") event failed! <font color=\"#FF0000\">**Blame!**</font> 😌\n";
                    matrixMessage += "The project [**" + project_name + "**](" + project_url + ") has failed.\n";
                }
                matrixMessage += "Pusher: *" + user + "*\tBranch: *" + pipeline_branch + "*\tCommit: [*" + commit_id.substring(-1, 8) + "*](" + commit_url + ")\tDuration: *" + duration_hra + "*";
                SendMessage(matrixRoom, matrixMessage);
            }
        } else if (event === "push") {
            LogInfo("push event");
            let matrixMessage = "";
            matrixMessage += "A new push was made to [" + repoName + "](https://github.com/" + repoName + ")  \n";

            for (commit of body.commits) {
                const user = commit.committer.name;
                const commit_url = commit.url;
                const commit_id = commit.id;
                const message = commit.message;

                matrixMessage += "[" + commit_id.substring(-1, 8) + "](" + commit_url + ") by " + user + " ```" + message + "```\n";
            }
            SendMessage(matrixRoom, matrixMessage);
        } else if (event === "ping") {
            LogInfo("Received GitHub ping");
            SendMessage(matrixRoom, "Yeah! I received a GitHub from *" + repoName + "* ping! It seems like everything is set up perfectly. 💖");
        } else {
            LogInfo("unwanted event " + event);
        }

        return res;
    });
}

function onMessage(data) {
    let res = Object.assign({}, data);
    if (!res.message.startsWith("!github") && !res.message.startsWith("!gh")) return;
    let args = res.message.split(" ");
    if (args.length <= 1) {
        res.response = "Hey there 👋🏼";
        return res
    }
    switch (args[1]) {
        case "test":
        case "ping":
            res.response = "Hey 👋 I'm here to help. 🧃";
            break;
        case "list":
            res.response = "I listen on the following repositories 🦻  \n";
            try {
                let rows = DBQuery("SELECT * FROM github WHERE room_id = ?", "string", data.roomID);
                if (rows.length === 0) {
                    res.response = "**I listen no none of your repositories** 😴";
                    break;
                }

                for (row of rows) {
                    res.response += "- [" + row[3] + "](https://github.com/" + row[3] + "/)\n";
                }
            } catch (e) {
                LogError("Whhops. Couldn't get data from database");
                res.response = "Something went wrong while fetching your data from the database 😭";
                break;
            }
            break;
        case "webhook":
            res.response = "Please use [this link](" + this.webServer.baseURL + "/github/webhook/" + SHA256(data.roomID) + ") as your webhook URL";
            break;
        case "add": {
            if (args.length !== 3) {
                res.response = "Huh? Have you missed to add the repository as argument?";
                break;
            }

            const repo = args[2].toLowerCase();

            if (repo.match(/([-_\w]+)\/([-_.\w]+)(?:#|@)?([-_.\w]+)?/) === null) {
                res.response = "Invalid GitHub repository given ❌";
                break;
            }

            try {
                DBPS("INSERT INTO github(room_id, room_hash, repo) VALUES (?, ?, ?)",
                    "string", data.roomID, "string", SHA256(data.roomID), "string", repo);
            } catch (e) {
                LogError("Whhops. Couldn't insert data into database");
                res.response = "Something went wrong while adding your repo to my watchlist 😭";
                break;
            }

            res.response = "I added " + repo + " to my watchlist ✅";
            break;
        }

        case "rm":
        case "remove": {
            if (args.length !== 3) {
                res.response = "Huh? Have you missed to add the repository as argument?";
                break;
            }

            const repo = args[2].toLowerCase();

            if (repo.match(/([-_\w]+)\/([-_.\w]+)(?:#|@)?([-_.\w]+)?/) === null) {
                res.response = "Invalid GitHub repository given ❌";
                break;
            }

            try {
                DBPS("DELETE FROM github WHERE room_id = ? AND repo = ?", "string", data.roomID, "string", repo);
            } catch (e) {
                LogError("Whhops. Couldn't delete data from database");
                res.response = "Something went wrong while removing your repo from my watchlist 😭";
                break;
            }

            res.response = "I removed " + repo + " from my watchlist 😪";
            break;
        }

    }
    return res;
}