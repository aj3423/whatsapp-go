import java.io.*;
import com.whatsapp.*;

public class A {
	public static void main(String[] args) {
		String name;
		String filename = "java.bin";

		try
		{
			FileInputStream file  = new FileInputStream(filename);
			ObjectInputStream out = new ObjectInputStream(file);


            TextData t = (TextData)out.readObject();

			System.out.println(t.backgroundColor);
			System.out.println(t.fontStyle);
			System.out.println(t.textColor);
			System.out.println(t.thumbnail);


			out.close();
			file.close();
		}
		catch(Exception e)
		{
			System.out.println("Exception: " + e.toString());
		}
	}
}
