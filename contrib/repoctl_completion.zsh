#compdef repoctl

function _repoctl_complete_actions {
	_values "repoctl actions" \
			"add[copy and add packages to the repository]" \
			"down[download and extract tarballs from AUR]" \
			"host[host repository on a network]" \
			"list[list packages that belong to the managed repository]" \
			"new[create a new repository or configuration file]" \
			"remove[remove package from repo]" \
			"reset[recreate repository database]" \
			"status[show pending changes and packages that can be upgraded]" \
			"update[update database in repository]" \
			"version[show version and date information]"
}

function _repoctl_complete_action_args {
	case $words[((CURRENT - 1))] in
		add)
			_files -g "*.pkg.tar.xz"
			;;
		remove)
			packages=($(repoctl list))
			compadd $packages
			;;
	esac
}

function _repoctl {
	_arguments  "1: :_repoctl_complete_actions" \
				"2: :_repoctl_complete_action_args" \
				{-b,--backup}"[backup obsolete files instead of deleting]" \
				{-B,--backup-dir}"[backup directory relative to repository path (default \"backup/\")]: :_files -/" \
				"--color[when to use color (auto|never|always) (default auto)]: :(auto never always)" \
				{-c,--columns}"[show items in columns rather than lines]" \
				"--debug[show unnecessary debugging information]" \
				{-q,--quiet}"[show minimal amount of information]"
}

_repoctl "$@"
