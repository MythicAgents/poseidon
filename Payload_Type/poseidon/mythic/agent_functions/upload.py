from mythic_payloadtype_container.MythicCommandBase import *
import json
import base64
import sys
from mythic_payloadtype_container.MythicRPC import *


class UploadArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="remote_path",
                display_name="Remote Path",
                type=ParameterType.String,
                description="Path where the uploaded file will be written.",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=2,
                        required=False
                    )
                ]
            ),
            CommandParameter(
                name="file_id",
                display_name="File to Upload",
                type=ParameterType.File,
                description="The file to be written to the remote path.",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=1
                    )
                ]
            ),
            CommandParameter(
                name="overwrite",
                display_name="Overwrite Exiting File",
                type=ParameterType.Boolean,
                description="Overwrite file if it exists.",
                default_value=False,
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=3,
                        required=False
                    )
                ]
             )
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class UploadCommand(CommandBase):
    cmd = "upload"
    needs_admin = False
    help_cmd = "upload"
    description = "upload a file to the target."
    version = 1
    supported_ui_features = ["file_browser:upload"]
    author = "@xorrior"
    argument_class = UploadArguments
    attackmapping = ["T1020", "T1030", "T1041", "T1105"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        try:
            file_resp = await MythicRPC().execute("get_file",
                                                  file_id=task.args.get_arg("file_id"),
                                                  task_id=task.id,
                                                  get_contents=False)
            if file_resp.status == MythicRPCStatus.Success:
                original_file_name = file_resp.response[0]["filename"]
            else:
                raise Exception("Error from Mythic: " + str(file_resp.error))
            if len(task.args.get_arg("remote_path")) == 0:
                task.args.add_arg("remote_path", original_file_name)
            elif task.args.get_arg("remote_path")[-1] == "/":
                task.args.add_arg("remote_path", task.args.get_arg("remote_path") + original_file_name)
            task.display_params = f"{original_file_name} to {task.args.get_arg('remote_path')}"
        except Exception as e:
            raise Exception("Error from Mythic: " + str(sys.exc_info()[-1].tb_lineno) + str(e))
        return task

    async def process_response(self, response: AgentResponse):
        pass
