#define NOTVALIDERR "One or both target players not found or not alive"

/*
 * RegisterCommands()
 * registers all commands to the level._commands list
 * using RegisterCommand() function
 * RegisterCommand
 * @name    string   name of the command
 * @minArgs int      minium amount of arguments required
 * @handler function the callback function
*/
RegisterCommands() {
    // Server side command for the backend to verify the game is ready
    RegisterCommand("plugin_ready", 0, ::on_ready);

    // Client commands
    RegisterCommand("swap",         2, ::swap);
    RegisterCommand("epilepsy",     2,  ::epilepsy);
    RegisterCommand("killplayer",   2, ::killplayer);
    RegisterCommand("hideplayer",   2, ::hideplayer);
    RegisterCommand("teleport",     2, ::teleport);
    RegisterCommand("setspectator", 2, ::setspectator);
    RegisterCommand("sayto",        3, ::sayto);
    RegisterCommand("giveweapon",   3, ::giveweapon);
    RegisterCommand("takeweapons",  2, ::takeweapons);
    RegisterCommand("freeze",       2, ::freezeplayer);
    RegisterCommand("setspeed",     3, ::setspeed);
    RegisterCommand("slapplayer",   2, ::slapplayer);
    RegisterCommand("loadout",      3, ::loadout);
    RegisterCommand("setgravity",   3, ::setgravity);
    RegisterCommand("dropgun",      2, ::dropgun);
    RegisterCommand("toggleleft",   0, ::toggleleft);
    RegisterCommand("bunnyhop",     2, ::bunnyhop);
    RegisterCommand("jumpheight",   3, ::jumpheight);
}

RegisterCommand(name, minArgs, handler) {
    if (!IsDefined(level._commands)) {
        level._commands = [];
    }

    cmd = SpawnStruct();
    cmd.name    = ToLower(name);
    cmd.minArgs = minArgs;
    cmd.handler = handler;

    level._commands[level._commands.size] = cmd;
}

ExecCommand(command) {
    if (!IsDefined(command) || command == "") {
        return;
    }

    parts = StrTok(command, " ");
    if (!IsDefined(parts) || parts.size == 0) {
        return;
    }

    def = FindRegisteredCommand(ToLower(parts[0]));
    if (!IsDefined(def)) {
        return;
    }

    args = [];
    for (i = 1; i < parts.size; i++) {
        args[args.size] = parts[i];
    }

    if (args.size < def.minArgs) {
        return;
    }
    
    thread [[def.handler]](args);
}

FindRegisteredCommand(name) {
    for (i = 0; i < level._commands.size; i++) {
        if (level._commands[i].name == name) {
            return level._commands[i];
        }
    }
    return undefined;
}

findPlayerByClientNum(n) {
    for ( i = 0; i < level.players.size; i++ ) {
        p = level.players[i];
        if ( p getEntityNumber() == n )
            return p;
    }
    return undefined;
}

on_ready(args) {
    scripts\mp\_plutoplugin_utils::SetOutDvar("success");
    wait 0.5;
    scripts\mp\_plutoplugin_utils::ResetOutDvar();
}

swap(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (args.size == 2) {
        if (scripts\mp\_plutoplugin_utils::isValidAndAlive(origin) || scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
            origin IPrintLnBold(NOTVALIDERR);
            return;
        }

        orgt = target GetOrigin();
        orgo = origin GetOrigin();

        origin SetOrigin(orgt);
        target SetOrigin(orgo);

        origin IPrintLnBold("You ^6swapped ^7with: ^6" + target.Name);
        target IPrintLnBold("You ^6swapped ^7with: ^6" + origin.Name);

        return;
    }

    other = findPlayerByClientNum(args[2]);
    if (scripts\mp\_plutoplugin_utils::isValidAndAlive(target) || scripts\mp\_plutoplugin_utils::isValidAndAlive(other)) {
        origin IPrintLnBold(NOTVALIDERR);
        return;
    }

    orgt = target GetOrigin();
    orgo = other GetOrigin();

    target SetOrigin(orgo);
    other SetOrigin(orgt);

    origin IPrintLnBold("Swapped ^6" + target.name + "^7 with: ^6" + other.name);
    origin IPrintLnBold("You ^6swapped ^7with: " + target.Name);
    target IPrintLnBold("You ^6swapped ^7with: " + other.Name);   
}

epilepsy(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!IsDefined(target)) {
        origin IPrintLnBold("Player ^6" + args[1] + " ^7not found");
        return;
    }

    target SetClientDvar("r_exposureTweak", "1");
    target SetClientDvar("r_exposureValue", "1");

    origin IPrintLnBold("Started epilepsy loop for ^6" + target.name);
    target IPrintLnBold("^6" + origin.name + "^7 started epilepsy loop for you");

    target endon("disconnect");
    for(;;) {
        target setClientDvar("r_exposureValue", "-3");
        wait 0.025;
        target playlocalsound("exp_barrel");
        wait 0.025;
        target playrumbleonentity("damage_heavy");
        wait 0.025;
        target setClientDvar("r_exposureValue", "16");
        wait 0.025;
    }
}

killplayer(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    target Suicide();
    origin IPrintLnBold("Killed ^6" + target.name);
    target IPrintLnBold("You got killed by ^6" + origin.name);
}

hideplayer(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    if (!IsDefined(target.is_hidden)) {
        target.is_hidden = false;
    }

    if (!target.is_hidden) {
        target Hide();
        target.is_hidden = true;
        origin IPrintLnBold("Hidden ^5" + target.name);

        if (origin.guid != target.guid) {
            target IPrintLnBold("^6" + origin.name + "^7 hid you");
        }
    } else {
        target Show();
        target.is_hidden = false;
        origin IPrintLnBold("Unhidden ^6" + target.name);

        if (origin.guid != target.guid) {
            target IPrintLnBold("^6" + origin.name + "^7 unhid you");
        }
    }
}

teleport(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (args.size == 2) {
        if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(origin) || !scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
            origin IPrintLnBold(NOTVALIDERR);
            return;
        }

        origin SetOrigin(target.origin);
        origin IPrintLnBold("Teleported to ^6" + target.name);
        return;
    }

    other = findPlayerByClientNum(args[2]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target) || !scripts\mp\_plutoplugin_utils::isValidAndAlive(other)) {
        origin IPrintLnBold(NOTVALIDERR);
        return;
    }

    target SetOrigin(other.origin);
    origin IPrintLnBold("Teleported ^6" + target.name + "^7 to ^6" + other.name);
    target IPrintLnBold("^6" + origin.name + " teleported you to ^6" + other.name);
}

setspectator(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target) || target.pers["team"] == "spectator") {
        origin IPrintLnBold("player ^6" + args[1] + "^7 already in spectator, or not found / alive");
        return;
    }

    target [[level.spectator]]();
    origin IPrintLnBold("Set ^6" + target.name + "^7 to spectator mode");
    target IPrintLnBold("^6" + target.name + "^7 set you to spectator");
}

sayto(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    msg = "";
    for (i = 2; i < args.size; i++) {
        if (i > 2) msg += " ";
        msg += args[i];
    }

    target IPrintLnBold("^6" + origin.name + "^7: " + msg);
    origin IPrintlnBold("Sent message to ^6" + target.name);
}

giveweapon(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    weapon = args[2];
    target GiveWeapon(weapon);
    target SwitchToWeapon(weapon);

    origin IPrintLnBold("Gave ^6" + target.name + "^7 " + weapon);
    if (origin.guid != target.guid) {
        target IPrintLnBold("^6" + origin.name + "^7 gave you weapon: ^6" + weapon);
    }
}

takeweapons(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    target TakeAllWeapons();
    origin IPrintLnBold("Took all weapons from ^6" + target.name);
    target IPrintLnBold("^6", origin.name, "^7 took all your weapons");
}

freezeplayer(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    if (!IsDefined(target.is_frozen)) {
        target.is_frozen = false;
    }

    if (!target.is_frozen) {
        target FreezeControls(true);
        target.is_frozen = true;
        origin IPrintLnBold("Frozen ^6" + target.name);

        if (origin.guid != target.guid) {
            target IPrintLnBold("^6" + origin.name + "^7 froze you");
        }
    } else {
        target FreezeControls(false);
        target.is_frozen = false;
        origin IPrintLnBold("Unfroze ^6" + target.name);

        if (origin.guid != target.guid) {
            target IPrintLnBold("^6" + origin.name + "^7 unfroze you");
        }
    }
}

setspeed(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    speed = float(args[2]);
    target SetMoveSpeedScale(speed);

    origin IPrintLnBold("Set ^6" + target.name + "^7 speed to ^6" + speed);
    if (origin.guid != target.guid) {
        target IPrintLnBold("^6" + origin.name + "^7 set your speed to ^6" + speed);
    }
}

slapplayer(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    vel = (RandomInt(400) - 100, RandomInt(700) - 100, 200);
    target SetVelocity(vel);

    origin IPrintLnBold("Slapped ^6" + target.name);
    if (origin.guid != target.guid) {
        target IPrintLnBold("^6" + origin.name + "^7 slapped you");
    }
}

loadout(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    weaponArg = ToLower(args[2]);
    target TakeAllWeapons();

    if (weaponArg == "ballista_mp" || weaponArg == "ballista" || weaponArg == "bal" || weaponArg == "1") {
        weapon = "ballista_mp+acog+steadyaim+extclip";
    } else if (weaponArg == "dsr50_mp" || weaponArg == "dsr50" || weaponArg == "dsr" || weaponArg == "2") {
        weapon = "dsr_mp+acog+steadyaim+extclip";
    } else {
        weapon = args[2];
    }

    target GiveWeapon(weapon, 0, RandomIntRange(1, 45));
    target SwitchToWeapon(weapon);

    origin IPrintLnBold("Gave ^6" + target.name + "^7 " + weapon);
    if (origin.guid != target.guid) {
        target IPrintLnBold("^6" + origin.name + "^7 gave you weapon " + weapon);
    }
}

setgravity(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    gravity = float(args[2]);
    target SetGravity(gravity);
    target SetClientDvar("bg_gravity", gravity);

    origin IPrintLnBold("Set ^6" + target.name + "^7 gravity to: ^6" + gravity);
    if (origin.guid != target.guid) {
        target IPrintLnBold("^6" + origin.name + "^7 set your gravity to: ^6" + gravity);
    }
}

dropgun(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    weapon = target GetCurrentWeapon();
    target DropItem(weapon);

    origin IPrintLnBold("Dropped ^6" + target.name + "^7 weapon");
    if (origin.guid != target.guid) {
        target IPrintLnBold("^6" + origin.name + "^7 dropped your weapon");
    }
}

toggleleft(args) {
    origin = findPlayerByClientNum(args[0]);
    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(origin)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    if (!IsDefined(origin.pers["left_toggled"])) {
        origin.pers["left_toggled"] = false;
    }

    if (!origin.pers["left_toggled"]) {
        origin SetClientDvar("cg_gun_y", "7");
        origin.pers["left_toggled"] = true;
    } else {
        origin SetClientDvar("cg_gun_y", "0");
        origin.pers["left_toggled"] = false;
    }
}

bunnyhop(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    if (!isDefined(target.pers["bunnyhop"])) {
        if (GetDvarInt("jump_slowdownEnabled") == 1) {
            target.pers["bunnyhop"] = false;
        } else {
            target.pers["bunnyhop"] = true;
        }
    }


    if (!target.pers["bunnyhop"]) {
        target SetClientDvar("sv_cheats", 1);
        target SetClientDvar("jump_slowdownEnable", 0);
        target SetClientDvar("sv_cheats", 0);
        target.pers["bunnyhop"] = true;

        origin IPrintLnBold("Bunnyhop enabled for ^6" + target.name);
        if (origin.guid != target.guid) {
            target IPrintLnBold("^6" + origin.name + "^7 enabled bunnyhop for you");
        }

    } else {
        target SetClientDvar("sv_cheats", 1);
        target SetClientDvar("jump_slowdownEnable", 1);
        target SetClientDvar("sv_cheats", 0);

        origin IPrintLnBold("Bunnyhop disabled for ^6" + target.name);
        if (origin.guid != target.guid) {
            target IPrintLnBold("^6" + origin.name + "^7 disabled bunnyhop for you");
        }
    }
}

jumpheight(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!scripts\mp\_plutoplugin_utils::isValidAndAlive(target)) {
        origin IPrintLnBold("player ^6" + args[1] + "^7 not alive or found");
        return;
    }

    height = float(args[2]);
    target SetClientDvar("sv_cheats", 1);
    target SetClientDvar("jump_height", height);
    target SetClientDvar("sv_cheats", 0);

    origin IPrintLnBold("set ^6" + target.name + "^7 jump height to ^6" + height);
    if (origin.guid != target.guid) {
        target IPrintLnBold("^6" + origin.name + "^7 set your jump height to ^6" + height);
    }
}