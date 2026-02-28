init() {
    level._dvars = SpawnStruct();
    level._dvars.enabled = "plutoplugin_enabled";
    level._dvars.inDvar  = "plutoplugin_in";
    level._dvars.outDvar = "plutoplugin_out";
    level._dvars.reset   = "";

    scripts\mp\_plutoplugin_utils::SetDvarIfUnitialized(level._dvars.enabled, 1);
    scripts\mp\_plutoplugin_utils::SetDvarIfUnitialized(level._dvars.inDvar,  "");
    scripts\mp\_plutoplugin_utils::SetDvarIfUnitialized(level._dvars.outDvar, "");

    level._commands = [];
    level thread inDvarListener();
}

inDvarListener() {
    level endon("game_ended");
    for(;;) {
        if (GetDvarInt(level._dvars.enabled) != 1) {
            wait 0.5;
            continue;
        }

        cmd = scripts\mp\_plutoplugin_utils::GetInDvar();
        if (cmd != "") {
            scripts\mp\_plutoplugin_utils::SetInDvar(level._dvars.reset);
            thread scripts\mp\_plutoplugin_commands::ExecCommand(cmd);
        }

        wait 0.01;
    }
}