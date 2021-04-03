const MAX_VALUE = 5 * 10 ^ 6;

function init() {

}

function getRndInteger(min, max) {
    return Math.floor(Math.random() * (max - min)) + min;
}

function onMessage(data) {
    let res = Object.assign({}, data);
    if (!res.message.startsWith("!dice")) return;
    let args = res.message.split(" ");

    if (args.length <= 1) {
        const rndDice = getRndInteger(1, 7);
        res.response = "I rolled a 🎲 with the number `" + rndDice + "` for you @" + data.sender;
        return res
    }

    if (args.length === 2 && args[1].match(/^-?\d+$/)) {
        const maxValue = parseInt(args[1]);
        if (maxValue > MAX_VALUE) {
        } else {
            const rndDice = getRndInteger(1, maxValue + 1 /* to include */);
            res.response = "I rolled a `" + rndDice + "` for you @" + data.sender;
        }
    } else if (args.length === 3 && args[1].match(/^-?\d+$/) && args[2].match(/^-?\d+$/)) {
        const minValue = parseInt(args[1]);
        const maxValue = parseInt(args[2]);
        if (maxValue > MAX_VALUE) {
            res.response = "***maxvalue* can not exceed `" + MAX_VALUE + "`";
        } else if (minValue < maxValue) {
            const rndDice = getRndInteger(minValue, maxValue + 1 /* to include */);
            res.response = "I rolled a `" + rndDice + "` for you @" + data.sender;
        } else {
            res.response = "**_minvalue_ needs to be smaller than _maxvalue_!**";
        }
    } else {
        res.response = "`!dice help` - prints this help\n";
        res.response += "`!dice` - rolls a dice with 6 sides\n";
        res.response += "`!dice [maxvalue]` - rolls a dice from 1 to maxvalue (included)\n";
        res.response += "`!dice [minvalue] [maxvalue]` - rolls a dice from minvalue (included) to maxvalue (included)";
    }
    return res;
}