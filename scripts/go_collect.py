import os
from pathlib import Path

def collect_go_code(base_dir, output_file):
    # 配置参数
    exclude_dirs = {'node_modules', '.git', 'dist', 'build', 'vendor'}
    exclude_files = {'go.sum', 'go.work.sum'}
    include_ext = {'.go', '.mod', '.yaml', '.yml'}  # 包含Go项目相关文件

    with open(output_file, 'w', encoding='utf-8') as out_f:
        # 生成目录结构说明
        out_f.write("=== GO PROJECT STRUCTURE ===\n")
        try:
            out_f.write(os.popen(f'tree -L 4 -I "node_modules|dist|build|vendor" {base_dir}').read() + "\n\n")
        except:
            out_f.write("# Note: Install 'tree' command for better structure visualization\n\n")

        # 遍历文件系统
        for root, dirs, files in os.walk(base_dir):
            # 跳过排除目录
            dirs[:] = [d for d in dirs if d not in exclude_dirs]

            for file in files:
                file_path = Path(root) / file
                ext = file_path.suffix.lower()

                if file in exclude_files:
                    continue
                if ext not in include_ext:
                    continue

                # 写入文件头（显示相对于项目根目录的路径）
                relative_path = file_path.relative_to(base_dir)
                header = f"\n// ====== FILE: {relative_path} ======\n\n"
                out_f.write(header)

                # 写入文件内容（带行号）
                try:
                    with open(file_path, 'r', encoding='utf-8') as in_f:
                        for i, line in enumerate(in_f, 1):
                            out_f.write(f"{i:4} | {line}")
                except UnicodeDecodeError:
                    out_f.write(f"// Binary file skipped: {relative_path}\n")
                except Exception as e:
                    out_f.write(f"// Error reading file: {str(e)}\n")

                out_f.write("\n\n")

if __name__ == '__main__':
    project_root = "/Users/123jiaru/Desktop/project/my/claimask"  # 硬编码项目路径
    output_path = os.path.join(project_root, "scripts", "go_project_code.txt")

    # 确保输出目录存在
    os.makedirs(os.path.dirname(output_path), exist_ok=True)

    collect_go_code(project_root, output_path)
    print(f"Go项目代码已整理到: {output_path}")