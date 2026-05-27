pm() {
  case "$1" in
    jump|switch|sw)
      local dir
      dir=$(command pm "$@" 2>&1)
      if [[ $? -eq 0 && -d "$dir" ]]; then
        cd "$dir"
      else
        echo "$dir" >&2
        return 1
      fi
      ;;
    *)
      command pm "$@"
      ;;
  esac
}
