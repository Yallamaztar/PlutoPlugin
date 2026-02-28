SetDvarIfUnitialized(dvar, value) {
    if (!isDefined(getDvar(dvar)) || getDvar(dvar) == "") {
        setDvar(dvar, value);
    }
}

GetDvarDefault(dvar, def) {
    d = getDvar(dvar);
    if (d != "") {
        return d;
    }

    return def;
}

isValidAndAlive(target) {
    return !isDefined(target) || IsAlive(target);
}

IsEnabled() {
    return GetDvarInt(level._dvars.enabled);
}

Disable() {
    SetDvar(level._dvars.enabled, 0);
}

Enable() {
    SetDvar(level._dvars.enabled, 1);
}

GetInDvar() {
    return GetDvar(level._dvars.inDvar);
}

SetInDvar(value) {
    SetDvar(level._dvars.inDvar, value);
}

GetOutDvar() {
    return GetDvar(level._dvars.outDvar);
}

SetOutDvar(value) {
    SetDvar(level._dvars.outDvar, value);
}

ResetOutDvar() {
    SetDvar(level._dvars.outDvar, "");
}