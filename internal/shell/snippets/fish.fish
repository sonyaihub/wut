function fish_command_not_found
    # If terminal-helper isn't on PATH, fall through cleanly.
    if not command -v terminal-helper >/dev/null 2>&1
        echo "fish: Unknown command: $argv[1]" >&2
        return 127
    end

    terminal-helper detect --line "$argv"
    set -l rc $status

    # Exit-code contract: 127 means "not natural language, let fish print its
    # normal command-not-found message". Any other code means terminal-helper
    # (or the harness it exec'd into) handled the line — propagate that code.
    if test $rc -eq 127
        echo "fish: Unknown command: $argv[1]" >&2
        return 127
    end
    return $rc
end
