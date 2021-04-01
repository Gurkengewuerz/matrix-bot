function init() {
    const alertWhen = ["success", "failure"];
    let tableStatement = "CREATE TABLE IF NOT EXISTS github (github_id INTEGER PRIMARY KEY AUTOINCREMENT, room_id TEXT NOT NULL,room_hash TEXT NOT NULL, repo TEXT NOT NULL);";
    try {
        DBPS(tableStatement);
    } catch (e) {
        LogError(e);
    }
    AddRoute("/github/webhook/{hash}", "POST", function callback(data) {
        LogInfo(JSON.stringify(data));
        let res = Object.assign({}, data);

        // HTTP_X_GITHUB_EVENT
        if((data.headers["HTTP_X_GITHUB_EVENT"] || "") !== "check_run") {
            LogInfo("Unwanted Event")
            return res;
        }

        res.statusCode = 201;
        res.response = "Hello " + data.params.hash;
        res.contentType = "text/plain";
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
            res.response = "I listen on the following repositories 🦻\n";
            try {
                let rows = DBQuery("SELECT * FROM github WHERE room_id = ?", "string", data.roomID);
                if(rows.length === 0) {
                    res.response = "**I listen no none of your repositories** 😴";
                    break;
                }
                for (row of rows) {
                    res.response += "- *" + row[3] + "*\n";
                }
            } catch (e) {
                LogError("Whhops. Couldn't get data from database");
                res.response = "Something went wrong while fetching your data from the database 😭";
                break;
            }
            break;
        case "webhook":
            res.response = "Please use " + this.webServer.baseURL + "/github/webhook/" + SHA256(data.roomID) + " as your webhook URL";
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