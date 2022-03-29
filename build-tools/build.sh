if ! output=$(svelte-kit build  2>&1); then 
    echo "$output"
    exit 2
fi
echo "Built Rill Developer!"