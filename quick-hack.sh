function quack() {
    HACK_HOME=${HACK_HOME:-$(pwd)}

    eval $($HACK_HOME/quick-hack ${@})
}