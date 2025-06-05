#/usr/bin/bash

arr=("auth_service" "tasks_service" "graphs_service" "groups_service" "api" "proto")

for w in ${arr[@]}; do
    rm -f $w.report
    find ../$w -type f -print0 | sort -z | while IFS= read -r -d '' file; do
        filename=$(basename -- "$file")
        if [[ "$filename" == *.* ]] || [[ "$filename" == "Dockerfile" ]]; then
            echo "$file"

            report_filename=$(echo "$file" | sed 's|^\.\./|project_root/|')

            echo -e "Файл: $report_filename\n" >>$w.report
            cat "$file" >>$w.report
            echo -e "\n\n" >>$w.report
        else
            echo -e "\n\tskip $file\n"
        fi
    done
done
