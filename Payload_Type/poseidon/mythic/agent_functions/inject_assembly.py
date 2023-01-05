import base64
import donut
import json
import os
import sys

from mythic_payloadtype_container.MythicCommandBase import *
from mythic_payloadtype_container.MythicRPC import *


class InjectAssemblyArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="file_id",
                display_name=".NET assembly to execute",
                type=ParameterType.File,
                description="Select the Bundle to execute in memory",
                parameter_group_info=[
                    ParameterGroupInfo(
                        required=True,
                        ui_position=1
                    )
                ]
            ),
            CommandParameter(
                name="args",
                display_name="Argument String",
                type=ParameterType.String,
                description=".NET command arguments",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=2
                    )
                ]
            ),
            CommandParameter(
                name="spawnas",
                display_name="spawnas process",
                default_value="rundll32.exe",
                type=ParameterType.String,
                description="sacrificial process to inject into",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=3
                    )
                ]
            )
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class InjectAssemblyCommand(CommandBase):
    cmd = "inject-assembly"
    needs_admin = False
    help_cmd = "inject-assembly"
    description = "Inject a .NET assembly into a sacrificial process"
    version = 1
    author = "@scottctaylor12"
    argument_class = InjectAssemblyArguments
    attributes = CommandAttributes(
        supported_os=[SupportedOS.Windows]
    )
    attackmapping = ["T1106"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        file_resp = await MythicRPC().execute("get_file",
                                              file_id=task.args.get_arg("file_id"),
                                              task_id=task.id,
                                              get_contents=True)
        if file_resp.status == MythicRPCStatus.Success:
            file_name = file_resp.response[0]["filename"]
            file_b64_contents = file_resp.response[0]["contents"]
        else:
            raise Exception("Error from Mythic: " + str(file_resp.error))

        # convert exe to shellcode with donut
        shellcode = convert_shellcode(file_name, file_b64_contents, task.args.get_arg("args"))

        await MythicRPC().execute("update_file",
                                  file_id=task.args.get_arg("file_id"),
                                  delete_after_fetch=True,
                                  contents=shellcode,
                                  comment="Uploaded shellcode into memory for execute_assembly")
        task.display_params = file_name + " with args: " + task.args.get_arg("args")
        return task

    async def process_response(self, response: AgentResponse):
        pass

'''
The uploaded file needs to be written to disk temporarily 
because the donut.create() function can only read files in 
to convert to shellcode instead of passing raw bytes :(
'''
def convert_shellcode(name, b64_bytes, params):
    # Write uploaded assembly to poseidon container disk
    save_location = "/tmp/" + name    
    with open(save_location , 'wb') as exe:
        exe.write(base64.b64decode(b64_bytes))

    # convert assembly on disk to shellcode
    if len(params) > 0:
        shellcode = donut.create(file=save_location, params=params)
    else:
        shellcode = donut.create(file=save_location)

    # cleanup
    dir = '/tmp'
    for f in os.listdir(dir):
        os.remove(os.path.join(dir, f))

    return shellcode
